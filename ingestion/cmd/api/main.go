package main

import (
	"database/sql"
	"log/slog"
	"os"

	"ingestion/internal/api"
	"ingestion/internal/recaptcha"
	"ingestion/internal/runiq"

	"github.com/wesleyskap/orkai-runiq/v3/queue"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Logger estruturado JSON conforme style-skills.md
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/bbb_development"
	}

	recaptchaKey := os.Getenv("RECAPTCHA_SECRET_KEY")
	if recaptchaKey == "" {
		recaptchaKey = "your_secret"
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		slog.Error("unable to open database connection", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	storage, err := queue.NewPostgresStorage(db)
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
