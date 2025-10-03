package routes

import (
	"myt-v/internal/db"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func SetupMediaRoutes(app *fiber.App, cfg AppConfig) {
	app.Get("/poster/:id", func(c *fiber.Ctx) error {
		var m db.Movie
		if err := db.DB.First(&m, "id = ?", c.Params("id")).Error; err != nil {
			return c.Status(404).SendString("not found")
		}

		movieDir := filepath.Dir(m.Path)
		poster := filepath.Join(movieDir, "image.webp")

		if !strings.HasPrefix(strings.ToLower(movieDir), strings.ToLower(filepath.Clean(cfg.Media))) {
			return c.Status(403).SendString("forbidden")
		}

		if _, err := os.Stat(poster); err == nil {
			return c.SendFile(poster)
		}
		return c.Status(404).SendString("poster not found")
	})
}
