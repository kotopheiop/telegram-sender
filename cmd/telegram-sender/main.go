package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"telegram-sender/pkg/logger"
	messageModule "telegram-sender/pkg/message"
	"telegram-sender/pkg/telegramApi"
	"time"
)

var (
	messageQueue = &messageModule.MessageQueue{}
	waitGroup    sync.WaitGroup
)

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		logger.Fatal("Токен бота не прописан в environment")
	}

	logger.Info("Сервис запустился на адресе localhost:8080")

	bot, err := telegramApi.InitBot(botToken)
	if err != nil {
		log.Fatal(err)
	}

	logger.Infof("Запущен бот с ником: %s (@%s)", bot.FirstName, bot.Username)

	messageSendingTicker := time.NewTicker(time.Second / 30)
	messageToSend := &messageModule.Message{}
	go func() {
		for range messageSendingTicker.C {
			messageToSend = messageQueue.Dequeue()
			if messageToSend != nil {
				err = bot.SendMessage(telegramApi.SendMessageParams{
					ChatID:          messageToSend.ChatID,
					Text:            messageToSend.Text,
					MessageThreadID: messageToSend.MessageThreadID,
				})

				if err != nil {
					logger.Errorf("Ошибка отправки сообщения: %v", err)
					retryAfter, _ := telegramApi.ParseRetryAfter(err)
					if retryAfter != 0 {
						logger.Warningf("Сервис прилёг отдохнуть на %d %s",
							retryAfter, telegramApi.Pluralize(retryAfter, "секунда", "секунды", "секунд"))
						time.Sleep(time.Duration(retryAfter) * time.Second)
						messageQueue.AddToTheBeginningEnqueue(*messageToSend) // Помещаем сообщение в начало очереди
					}
				} else {
					logger.Info("Сообщение успешно отправлено")
				}
			}
		}
	}()

	http.HandleFunc("/send", sendMessageHandler)

	// Проверка работоспособности
	http.HandleFunc("/health", healthCheckerHandler)

	logger.Fatal(http.ListenAndServe(":8080", nil))
	waitGroup.Wait()
}

func healthCheckerHandler(w http.ResponseWriter, r *http.Request) {
	requestDateTime := time.Now().Format(logger.DateTimeFormat)
	logger.Info("Проверка работоспособности")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(requestDateTime))
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {

	requestDateTine := time.Now().Format(logger.DateTimeFormat)

	chatId, err := strconv.ParseInt(r.Header.Get("ChatID"), 10, 64)
	if err != nil {
		logger.Errorf("Ошибка чтения ChatID: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	threadId, _ := strconv.Atoi(r.Header.Get("MessageThreadID"))

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		logger.Errorf("Ошибка чтения тела запроса: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newMessage := messageModule.Message{
		ChatID:          chatId,
		Text:            fmt.Sprintf("%s\n%s", requestDateTine, string(body)),
		MessageThreadID: &threadId,
	}

	messageQueue.Enqueue(newMessage) // Добавляем сообщение в очередь

	runtime.GC() //Вызовем сборщик мусора
}
