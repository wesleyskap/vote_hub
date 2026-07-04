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
	// Logger estruturado JSON com nível DEBUG para máxima rastreabilidade
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	slog.Debug("worker initializing", "component", "ingestion-worker")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := mustOpenDB()
	defer db.Close()

	// RedisStorage concreto — implementa WorkerPoolStorage, ClientStorage e ServerStorage
	storage := mustOpenRedisStorage()

	aggBuffer := vote.NewAggregationBuffer(db, 5*time.Second, 500)
	go aggBuffer.Start(ctx)

	voteJob := vote.NewVoteJob(db, aggBuffer)
	workerPool := buildWorkerPool(storage, voteJob)

	startDashboard(storage)

	slog.Info("starting Runiq Worker Pool for votes with autoscaling",
		"minConcurrency", 15,
		"maxConcurrency", 100,
		"aggFlushInterval", "5s",
		"aggFlushSize", 500,
	)

	if err := workerPool.Start(ctx, "votes_queue"); err != nil {
		slog.Error("worker pool stopped with error", "err", err)
	}

	slog.Info("worker pool stopped cleanly")
}

// mustOpenDB abre a conexão com o banco e configura o pool de conexões.
// Encerra o processo em caso de erro sem banco o worker não tem utilidade.
func mustOpenDB() *sql.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/bbb_development"
		// WARN: ausencia de DATABASE_URL em produção indica misconfiguration
		slog.Warn("DATABASE_URL not set, falling back to localhost default")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("worker unable to open database connection", "err", err)
		os.Exit(1)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)

	if err := db.Ping(); err != nil {
		// WARN: ping falhou e o banco pode estar subindo ainda.. o worker tentará novamente no primeiro job
		slog.Warn("database ping failed on startup — will retry on first job", "err", err)
	} else {
		slog.Debug("database connection established successfully")
	}

	return db
}

// mustOpenRedisStorage inicializa o RedisStorage concreto do Runiq.
// Retorna o tipo concreto pois precisa satisfazer WorkerPoolStorage e ServerStorage.
// Encerra o processo em caso de falha — sem Redis a fila não funciona.
func mustOpenRedisStorage() *queue.RedisStorage {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		PoolSize: 1000,
	})

	storage, err := queue.NewRedisStorage(redisClient)
	if err != nil {
		slog.Error("unable to initialize redis storage", "err", err)
		os.Exit(1)
	}

	slog.Debug("redis storage initialized successfully")
	return storage
}

// buildWorkerPool configura o pool com autoscaling dinâmico e leader election.
func buildWorkerPool(storage queue.WorkerPoolStorage, voteJob *vote.VoteJob) *runiq.WorkerPool {
	dynConfig := queue.DynamicConcurrencyConfig{
		CheckInterval:   5 * time.Second,
		MinConcurrency:  15,
		MaxConcurrency:  100,
		QueueDepthLimit: 20,
		ScaleUpStep:     10,
		ScaleDownStep:   2,
	}

	slog.Debug("worker pool dynamic concurrency configured",
		"minConcurrency", dynConfig.MinConcurrency,
		"maxConcurrency", dynConfig.MaxConcurrency,
		"checkInterval", dynConfig.CheckInterval,
		"queueDepthLimit", dynConfig.QueueDepthLimit,
	)

	workerPool := runiq.NewWorkerPool(
		storage,
		dynConfig.MinConcurrency,
		queue.WithDynamicConcurrency(dynConfig),
		queue.WithLeaderElection(30*time.Second),
	)
	workerPool.Register("process_vote", voteJob)
	return workerPool
}

// startDashboard inicia o servidor HTTP do painel Runiq em goroutine separada.
func startDashboard(storage *queue.RedisStorage) {
	go func() {
		dashboard := queue.NewServer(storage, ":8081")
		slog.Info("starting Runiq Dashboard", "port", 8081)
		if err := dashboard.Start(); err != nil {
			slog.Error("runiq dashboard error", "err", err)
		}
	}()
}
