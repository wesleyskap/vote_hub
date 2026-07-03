package main

import (
	"database/sql"
	"log/slog"
	"os"

	"ingestion/internal/api"
	"ingestion/internal/recaptcha"
	"ingestion/internal/runiq"

	"github.com/redis/go-redis/v9"
	"github.com/wesleyskap/orkai-runiq/v3/queue"

	_ "github.com/lib/pq"
)

func main() {
	// Logger estruturado JSON conforme style-skills.md
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/bbb_development"
	}

	recaptchaKey := os.Getenv("RECAPTCHA_SECRET_KEY")
	if recaptchaKey == "" {
		recaptchaKey = "your_secret"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("unable to open database connection", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		PoolSize: 1000,
	})
	storage, err := queue.NewRedisStorage(redisClient)
	if err != nil {
		slog.Error("unable to initialize postgres storage", "err", err)
		os.Exit(1)
	}

	// Injeção de dependências explícita em main()
	enqueuer := runiq.NewClient(storage)
	verifier := recaptcha.NewGoogleVerifier(recaptchaKey)
	ingester := api.NewVoteIngester(enqueuer, verifier)

	app := api.SetupRouter(ingester)

	slog.Info("starting Vote Ingestion API", "port", 8080)
	if err := app.Listen(":8080"); err != nil {
		slog.Error("server shutdown failed", "err", err)
	}
}
