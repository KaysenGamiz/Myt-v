package routes

import (
	"myt-v/internal/db"

	"github.com/gofiber/fiber/v2"
)

func SetupCatalogRoutes(app *fiber.App) {
	app.Get("/catalog", func(c *fiber.Ctx) error {
		var movies []db.Movie
		db.DB.Find(&movies)
		return c.JSON(movies)
	})
}
