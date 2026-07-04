package vote

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// VoteJob processa um único voto persistindo-o no banco e registrando
// o incremento no AggregationBuffer para flush periódico.
// Campos ordenados por tamanho para minimizar padding.
type VoteJob struct {
	db     *sql.DB            // 8 bytes
	buffer *AggregationBuffer // 8 bytes
}

// NewVoteJob cria uma nova instância de VoteJob com injeção de dependências explícita.
// Exemplo de uso:
//
//	job := vote.NewVoteJob(db, aggBuffer)
func NewVoteJob(db *sql.DB, buffer *AggregationBuffer) *VoteJob {
	return &VoteJob{db: db, buffer: buffer}
}

// Perform insere o voto no banco e registra o incremento no buffer de agregação.
// Logs estruturados emitidos em todos os níveis (Debug, Warn, Error) para rastreabilidade.
// Exemplo de uso:
//
//	err := job.Perform(ctx, []byte(`{"paredao_id":1,"participant_id":2,"fingerprint_id":"x"}`))
func (v *VoteJob) Perform(ctx context.Context, args []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var payload Payload
	if err := json.Unmarshal(args, &payload); err != nil {
		// ERROR: payload malformado — indica bug no enqueuer ou mensagem corrompida no Redis
		slog.Error("vote job received malformed payload",
			"err", err,
			"raw_payload", string(args),
		)
		return fmt.Errorf("failed to unmarshal vote payload (offending: %s): %w", string(args), err)
	}

	if payload.ParedaoID == 0 || payload.ParticipantID == 0 {
		// ERROR: IDs obrigatórios ausentes — situação de erro garantida para o desafio
		slog.Error("vote job received payload with missing required IDs",
			"trace_id", payload.TraceID,
			"paredao_id", payload.ParedaoID,
			"participant_id", payload.ParticipantID,
			"fingerprint_id", payload.Fingerprint,
		)
		return fmt.Errorf("invalid payload (expected non-zero ids, got paredao_id: %d, participant_id: %d)",
			payload.ParedaoID, payload.ParticipantID)
	}

	slog.Debug("processing vote job",
		"trace_id", payload.TraceID,
		"paredao_id", payload.ParedaoID,
		"participant_id", payload.ParticipantID,
		"fingerprint_id", payload.Fingerprint,
	)

	now := time.Now()
	if err := v.insertVote(ctx, payload, now); err != nil {
		return err
	}

	// Incrementa em memória — sem roundtrip ao banco.
	v.buffer.Add(payload.ParedaoID, payload.ParticipantID, now.Truncate(time.Hour))

	slog.Info("vote processed successfully",
		"trace_id", payload.TraceID,
		"paredao_id", payload.ParedaoID,
		"participant_id", payload.ParticipantID,
	)

	return nil
}

// insertVote executa o INSERT atômico na tabela votes.
// Separado do Perform para manter funções abaixo de 20 linhas.
func (v *VoteJob) insertVote(ctx context.Context, payload Payload, now time.Time) error {
	const query = `
		INSERT INTO votes (paredao_id, participant_id, fingerprint_id, created_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := v.db.ExecContext(ctx, query,
		payload.ParedaoID, payload.ParticipantID, payload.Fingerprint, now,
	); err != nil {
		slog.Error("failed to insert vote into database",
			"trace_id", payload.TraceID,
			"paredao_id", payload.ParedaoID,
			"participant_id", payload.ParticipantID,
			"err", err,
		)
		return fmt.Errorf("failed to insert vote into DB: %w", err)
	}
	return nil
}
