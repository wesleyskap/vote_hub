package api

import (
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// SetupRouter inicializa a API de Ingestão configurando o Rate Limiter
func SetupRouter(ingester *VoteIngester) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ProxyHeader:           fiber.HeaderXForwardedFor,
	})

	// Habilita CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	prometheus := fiberprometheus.New("bbb_ingestion_api")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// Rate Limiter: Bloqueia IPs que excedem 10 requisições por segundo
	app.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Second,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "too many votes from this IP, slow down",
			})
		},
	}))

	app.Post("/api/v1/votes", ingester.Ingest)

	return app
}
