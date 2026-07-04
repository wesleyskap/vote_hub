package vote

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"
)

// aggKey identifica unicamente uma linha de agregação por hora.
// Campos ordenados por tamanho para minimizar padding.
type aggKey struct {
	VoteHour      time.Time // 24 bytes
	ParedaoID     int64     // 8 bytes
	ParticipantID int64     // 8 bytes
}

// AggregationBuffer acumula contagens de votos em memória e realiza flush
// periódico ao banco, eliminando lock contention no UPDATE da mesma linha.
//
// Trade-off: agregados podem ter até FlushInterval de atraso — aceitável
// para um placar ao vivo onde consistência eventual é suficiente.
type AggregationBuffer struct {
	db     *sql.DB          // 8 bytes
	counts map[aggKey]int64 // 8 bytes

	mu            sync.Mutex    // 8 bytes
	flushInterval time.Duration // 8 bytes
	flushSize     int           // 8 bytes
}

// NewAggregationBuffer cria o buffer com flush periódico e por volume.
// flushInterval: intervalo máximo entre flushes (ex: 5s).
// flushSize:     flush antecipado ao atingir N chaves distintas (ex: 500).
// Exemplo de uso:
//
//	buf := vote.NewAggregationBuffer(db, 5*time.Second, 500)
//	go buf.Start(ctx)
func NewAggregationBuffer(db *sql.DB, flushInterval time.Duration, flushSize int) *AggregationBuffer {
	return &AggregationBuffer{
		db:            db,
		counts:        make(map[aggKey]int64),
		flushInterval: flushInterval,
		flushSize:     flushSize,
	}
}

// Add incrementa o contador em memória para a chave (paredao, participant, hora).
// Dispara flush assíncrono se o buffer atingir o limite de tamanho.
func (b *AggregationBuffer) Add(paredaoID, participantID int64, voteHour time.Time) {
	key := aggKey{voteHour, paredaoID, participantID}
	b.mu.Lock()
	b.counts[key]++
	shouldFlush := len(b.counts) >= b.flushSize
	b.mu.Unlock()

	if shouldFlush {
		// Flush assíncrono para não bloquear o caller (worker goroutine)
		go b.Flush(context.Background())
	}
}

// Flush persiste os contadores acumulados com um único upsert por chave.
// Faz swap atômico do mapa para minimizar o tempo de lock.
func (b *AggregationBuffer) Flush(ctx context.Context) {
	snapshot := b.swapCounts()
	if len(snapshot) == 0 {
		return
	}

	start := time.Now()
	b.persistSnapshot(ctx, snapshot)
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		// WARN: flush lento pode indicar pressão no banco ou lock contention residual
		slog.Warn("aggregation flush took longer than expected",
			"duration_ms", elapsed.Milliseconds(),
			"keys_flushed", len(snapshot),
		)
		return
	}

	slog.Info("aggregation flush done",
		"keys_flushed", len(snapshot),
		"duration_ms", elapsed.Milliseconds(),
	)
}

// Start executa o loop de flush periódico. Deve ser chamado em goroutine separada.
// Ao receber ctx.Done(), realiza um flush final antes de encerrar.
func (b *AggregationBuffer) Start(ctx context.Context) {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	slog.Debug("aggregation buffer started", "flush_interval", b.flushInterval, "flush_size", b.flushSize)

	for {
		select {
		case <-ctx.Done():
			slog.Info("aggregation buffer shutting down — performing final flush")
			b.Flush(context.Background())
			return
		case <-ticker.C:
			b.Flush(ctx)
		}
	}
}

// swapCounts substitui atomicamente o mapa interno por um novo vazio,
// retornando o snapshot para processamento fora do lock.
func (b *AggregationBuffer) swapCounts() map[aggKey]int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.counts) == 0 {
		return nil
	}
	snapshot := b.counts
	b.counts = make(map[aggKey]int64, len(snapshot))
	return snapshot
}

// persistSnapshot executa um upsert por chave única no banco de dados.
// Erros de persistência são logados mas não interrompem os demais upserts.
func (b *AggregationBuffer) persistSnapshot(ctx context.Context, snapshot map[aggKey]int64) {
	const query = `
		INSERT INTO vote_aggregations_by_hours (paredao_id, participant_id, vote_hour, total_votes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (paredao_id, participant_id, vote_hour)
		DO UPDATE SET total_votes = vote_aggregations_by_hours.total_votes + EXCLUDED.total_votes
	`
	for key, count := range snapshot {
		if _, err := b.db.ExecContext(ctx, query,
			key.ParedaoID, key.ParticipantID, key.VoteHour, count,
		); err != nil {
			// ERROR: falha ao persistir — contagem pode ser perdida se o pod reiniciar
			slog.Error("failed to persist vote aggregation batch",
				"paredao_id", key.ParedaoID,
				"participant_id", key.ParticipantID,
				"vote_hour", key.VoteHour,
				"count_lost", count,
				"err", err,
			)
		}
	}
}
