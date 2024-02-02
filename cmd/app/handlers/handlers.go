package handlers

import (
	"fmt"
	"strconv"
	"telegram-sender/pkg/logger"
	messageModule "telegram-sender/pkg/message"
	"time"

	"github.com/gofiber/fiber/v2"
)

func HealthCheckerHandler(c *fiber.Ctx) error {
	requestDateTime := time.Now().Format(logger.DateTimeFormat)
	logger.Info("Проверка работоспособности")

	return c.JSON(fiber.Map{
		"requestDateTine": requestDateTime,
	})
}

func SendMessageHandler(c *fiber.Ctx, messageQueue *messageModule.Queue) error {
	requestDateTime := time.Now().Format(logger.DateTimeFormat)

	chatId, err := strconv.ParseInt(c.Get("ChatID"), 10, 64)
	if err != nil {
		logger.Errorf("Ошибка чтения ChatID: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString("Ошибка чтения ChatID")
	}
	threadId, _ := strconv.Atoi(c.Get("MessageThreadID"))

	body := c.Body()
	if len(body) == 0 {
		logger.Errorf("Ошибка чтения тела запроса: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString("Ошибка чтения тела запроса")
	}

	newMessage := messageModule.Message{
		ChatID:          chatId,
		Text:            fmt.Sprintf("%s\n%s", requestDateTime, string(body)),
		MessageThreadID: &threadId,
	}

	messageQueue.Enqueue(newMessage) // Добавляем сообщение в очередь

	return c.JSON(fiber.Map{
		"requestDateTime": requestDateTime,
		"queueSize":       messageQueue.Size(),
	})
}
