package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"telegramSender/utils/logger"
	"telegramSender/utils/telegramApi"
	"time"
)

var wg sync.WaitGroup

type Message struct {
	ChatID          int64
	MessageThreadID *int
	Text            string
}

type MessageQueue struct {
	sync.Mutex
	messages []Message
}

func (q *MessageQueue) AddToTheBeginningEnqueue(message Message) {
	q.Lock()
	defer q.Unlock()
	q.messages = append([]Message{message}, q.messages...)
}

func (q *MessageQueue) Enqueue(message Message) {
	q.Lock()
	defer q.Unlock()
	q.messages = append(q.messages, message)
}

func (q *MessageQueue) Dequeue() *Message {
	q.Lock()
	defer q.Unlock()
	if len(q.messages) == 0 {
		return nil
	}
	message := q.messages[0]
	q.messages = q.messages[1:]
	return &message
}

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		logger.L.Fatal("Токен бота не прописан в environment")
	}

	logger.L.Info("Сервис запустился на адресе localhost:8080")

	bot, err := telegramApi.NewBot(botToken)
	if err != nil {
		log.Fatal(err)
	}

	logger.L.Infof("Запущен бот с ником: %s (@%s)", bot.FirstName, bot.Username)

	messageQueue := &MessageQueue{}
	ticker := time.NewTicker(time.Second / 30)
	message := &Message{}
	go func() {
		for range ticker.C {
			message = messageQueue.Dequeue()
			if message != nil {
				err = bot.SendMessage(telegramApi.SendMessageParams{
					ChatID:          message.ChatID, // ID чата
					Text:            message.Text,
					MessageThreadID: message.MessageThreadID,
				})

				if err != nil {
					logger.L.Errorf("Ошибка отправки сообщения: %v", err)
					retryAfter, _ := parseRetryAfter(err)
					if retryAfter != 0 {
						logger.L.Warningf("Сервис прилёг отдохнуть на %d %s",
							retryAfter, pluralize(retryAfter, "секунда", "секунды", "секунд"))
						time.Sleep(time.Duration(retryAfter) * time.Second)
						messageQueue.AddToTheBeginningEnqueue(*message) // Помещаем сообщение в начало очереди
					}
				} else {
					logger.L.Info("Сообщение успешно отправлено")
				}
			}
		}
	}()

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {

		requestDateTine := time.Now().Format(logger.DateTimeFormat)

		chatId, err := strconv.ParseInt(r.Header.Get("ChatID"), 10, 64)
		if err != nil {
			logger.L.Errorf("Ошибка чтения ChatID: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		threadId, _ := strconv.Atoi(r.Header.Get("MessageThreadID"))

		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			logger.L.Errorf("Ошибка чтения тела запроса: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		message := Message{
			ChatID:          chatId,
			Text:            fmt.Sprintf("%s\nСообщение:\n%s", requestDateTine, string(body)),
			MessageThreadID: &threadId,
		}
		messageQueue.Enqueue(message) // Добавляем сообщение в очередь

		runtime.GC() //Вызовем сборщик мусора
	})

	// Проверка работоспособности
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		requestDateTime := time.Now().Format(logger.DateTimeFormat)
		logger.L.Info("Проверка работоспособности")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(requestDateTime))
	})

	logger.L.Fatal(http.ListenAndServe(":8080", nil))
	wg.Wait()
}

func parseRetryAfter(err error) (int, error) {
	re := regexp.MustCompile(`retry after (\d+)`)
	matches := re.FindStringSubmatch(err.Error())
	if len(matches) < 2 {
		return 0, fmt.Errorf("no retry after in error message")
	}
	retryAfter, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert retry after to int: %w", err)
	}
	return retryAfter, nil
}

func pluralize(n int, singular, plural1, plural2 string) string {
	n = n % 100
	if n > 10 && n < 20 {
		return plural2
	}
	n = n % 10
	if n == 1 {
		return singular
	}
	if n > 1 && n < 5 {
		return plural1
	}
	return plural2
}
