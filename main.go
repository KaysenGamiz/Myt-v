package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"myt-v/internal/db"
	"myt-v/internal/scanner"
	"myt-v/internal/stream"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type AppConfig struct {
	Bind    string
	Media   string
	HLSDir  string
	EnvName string
}

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
	cfg := AppConfig{
		Bind:    getEnv("BIND", "127.0.0.1:8080"), // local; ideally behind VPN
		Media:   getEnv("MEDIA_DIR", "D:\\Peliculas"),
		HLSDir:  getEnv("HLS_DIR", "./hls"),
		EnvName: getEnv("APP_ENV", "dev"),
	}

	ensureDir(cfg.Media)
	ensureDir(cfg.HLSDir)
	ensureDir("./public")

	// Inicializar DB (SQLite local)
	db.InitDB("catalog.db")

	// Escanear carpeta de películas al arranque
	scanner.ScanDir(cfg.Media)

	// Init Fiber
	app := fiber.New(fiber.Config{
		AppName: "Myt-V",
	})

	// Middlewares
	app.Use(recover.New())
	app.Use(logger.New())

	// Healthcheck
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Config
	app.Get("/config", func(c *fiber.Ctx) error {
		return c.JSON(cfg)
	})

	// Catalog from SQLite
	app.Get("/catalog", func(c *fiber.Ctx) error {
		var movies []db.Movie
		db.DB.Find(&movies)
		return c.JSON(movies)
	})

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
		return c.SendFile("./public/watch.html")
	})

	// Estáticos: UI + HLS
	app.Static("/", "./public")
	app.Static("/hls", cfg.HLSDir)

	// Arranque y apagado limpio
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
