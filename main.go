package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"myt-v/internal/db"
	"myt-v/internal/scanner"
	"myt-v/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func ensureDir(path string) {
	if path == "" {
		return
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		log.Fatalf("Can't create dir %s: %v", path, err)
	}
}

func main() {
	cfg := routes.AppConfig{
		Bind:    getEnv("BIND", "0.0.0.0:8080"),
		Media:   getEnv("MEDIA_DIR", "D:\\Peliculas"),
		HLSDir:  getEnv("HLS_DIR", "./hls"),
		EnvName: getEnv("APP_ENV", "dev"),
	}

	ensureDir(cfg.Media)
	ensureDir(cfg.HLSDir)
	ensureDir("./public")

	db.InitDB("catalog.db")

	scanner.ScanDir(cfg.Media)

	app := fiber.New(fiber.Config{
		AppName: "Myt-V",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	routes.SetupRoutes(app, cfg)

	go func() {
		log.Printf("Howdy partner! Myt-V (%s) listening on http://%s", cfg.EnvName, cfg.Bind)
		if err := app.Listen(cfg.Bind); err != nil {
			log.Fatalf("error Listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	_ = app.Shutdown()
	log.Println("Done. See ya!")
}
