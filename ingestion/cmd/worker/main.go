package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"time"

	"ingestion/internal/runiq"
	"ingestion/internal/vote"

	"github.com/redis/go-redis/v9"
	"github.com/wesleyskap/orkai-runiq/v3/queue"

	_ "github.com/lib/pq"
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

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("worker unable to open database connection", "err", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)

	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		PoolSize: 1000,
	})
	storage, err := queue.NewRedisStorage(redisClient)
	if err != nil {
		slog.Error("unable to initialize redis storage", "err", err)
		os.Exit(1)
	}

	// Buffer de agregação: flush a cada 5s ou quando acumular 500 chaves únicas
	aggBuffer := vote.NewAggregationBuffer(db, 5*time.Second, 500)
	go aggBuffer.Start(ctx)

	voteJob := vote.NewVoteJob(db, aggBuffer)

	// Configuração do dynamic concurrency (autoscaling)
	// Começa em 15, escala até 100, checa a cada 5 segundos
	dynConfig := queue.DynamicConcurrencyConfig{
		CheckInterval:   5 * time.Second,
		MinConcurrency:  15,
		MaxConcurrency:  100,
		QueueDepthLimit: 20,
		ScaleUpStep:     10,
		ScaleDownStep:   2,
	}

	// Inicializa o WorkerPool com autoscaling e leader election
	workerPool := runiq.NewWorkerPool(
		storage,
		dynConfig.MinConcurrency,
		queue.WithDynamicConcurrency(dynConfig),
		queue.WithLeaderElection(30*time.Second),
	)
	workerPool.Register("process_vote", voteJob)

	// Iniciar Dashboard do Runiq
	go func() {
		dashboard := queue.NewServer(storage, ":8081")
		slog.Info("starting Runiq Dashboard", "port", 8081)
		if err := dashboard.Start(); err != nil {
			slog.Error("runiq dashboard error", "err", err)
		}
	}()

	slog.Info("starting Runiq Worker Pool for votes with autoscaling",
		"minConcurrency", dynConfig.MinConcurrency,
		"maxConcurrency", dynConfig.MaxConcurrency,
		"aggFlushInterval", "5s",
		"aggFlushSize", 500,
	)
	if err := workerPool.Start(ctx, "votes_queue"); err != nil {
		slog.Error("worker pool stopped with error", "err", err)
	}

	slog.Info("worker pool stopped cleanly")
}
