package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupHealthRoutes(app *fiber.App, cfg AppConfig) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/config", func(c *fiber.Ctx) error {
		return c.JSON(cfg)
	})
}
