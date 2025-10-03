package routes

import (
	"log"
	"myt-v/internal/db"
	"myt-v/internal/stream"

	"github.com/gofiber/fiber/v2"
)

func SetupStreamRoutes(app *fiber.App, cfg AppConfig) {
	app.Get("/stream/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var m db.Movie
		if err := db.DB.First(&m, "id = ?", id).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "movie not found"})
		}
		m3u8, err := stream.StartHLS(c.Context(), m, stream.HLSConfig{HLSDir: cfg.HLSDir})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"m3u8": m3u8})
	})

	app.Get("/watch/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		log.Printf("Opening watch view for id=%s", id)
		return c.SendFile("./views/watch.html")
	})
}
