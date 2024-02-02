package server

import (
	"github.com/gofiber/fiber/v2"
	"telegram-sender/cmd/app/handlers"
	"telegram-sender/pkg/logger"
	messageModule "telegram-sender/pkg/message"
)

func StartServer(messageQueue *messageModule.Queue) {
	app := fiber.New(fiber.Config{
		ServerHeader: "Fiber",
		AppName:      "Telegram Sender App v1.0.1",
	})

	apiRouter := app.Group("/api")
	apiRouter.Get("/health", handlers.HealthCheckerHandler)
	apiRouter.Post("/send", func(c *fiber.Ctx) error {
		return handlers.SendMessageHandler(c, messageQueue)
	})

	logger.Fatal(app.Listen(":8080"))
}
