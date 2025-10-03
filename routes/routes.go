package routes

import (
	"github.com/gofiber/fiber/v2"
)

type AppConfig struct {
	Bind    string
	Media   string
	HLSDir  string
	EnvName string
}

func SetupRoutes(app *fiber.App, cfg AppConfig) {
	SetupHealthRoutes(app, cfg)
	SetupCatalogRoutes(app)
	SetupMediaRoutes(app, cfg)
	SetupStreamRoutes(app, cfg)

	app.Static("/", "./public")
	app.Static("/", "./views")
	app.Static("/hls", cfg.HLSDir)
}
