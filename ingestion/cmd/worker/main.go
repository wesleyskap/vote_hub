package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ingestion/internal/runiq"
	"ingestion/internal/vote"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		slog.Error("worker unable to open database connection", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	storage, err := queue.NewPostgresStorage(db)
	if err != nil {
		slog.Error("unable to initialize postgres storage", "err", err)
		os.Exit(1)
	}

	voteJob := vote.NewVoteJob(db)

	// Inicializa o WorkerPool do runiq local (que encapsula o oficial)
	workerPool := runiq.NewWorkerPool(storage, 500)
	workerPool.Register("process_vote", voteJob)

	slog.Info("starting Runiq Worker Pool for votes", "concurrency", 500)
	if err := workerPool.Start(ctx, "votes_queue"); err != nil {
		slog.Error("worker pool stopped with error", "err", err)
	}

	slog.Info("worker pool stopped cleanly")
}
