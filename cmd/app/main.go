package main

import (
	"errors"
	"os"
	"sync"
	"telegram-sender/cmd/app/server"
	"telegram-sender/pkg/logger"
	messageModule "telegram-sender/pkg/message"
	"telegram-sender/pkg/telegramApi"
	"time"
)

var (
	messageQueue = &messageModule.Queue{}
	waitGroup    sync.WaitGroup
)

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		logger.Fatal("Токен бота не прописан в environment")
	}

	bot, err := telegramApi.InitBot(botToken)
	if err != nil {
		logger.Fatal(err)
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
					var telegramErr *telegramApi.TelegramError
					if errors.As(err, &telegramErr) {
						retryAfter, _ := telegramApi.ParseRetryAfter(telegramErr)
						if retryAfter != 0 {
							logger.Warningf("Сервис прилёг отдохнуть на %d %s",
								retryAfter, telegramApi.Pluralize(retryAfter, "секунда", "секунды", "секунд"))
							time.Sleep(time.Duration(retryAfter) * time.Second)
							messageQueue.AddToTheBeginningEnqueue(*messageToSend) // Помещаем сообщение в начало очереди
						}
					}
				} else {
					logger.Info("Сообщение успешно отправлено")
				}
			}
		}
	}()

	server.StartServer(messageQueue)

	waitGroup.Wait()
}
