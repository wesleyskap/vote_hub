package vote

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// VoteJob processa um único voto de forma transacional e atômica.
// A agregação horária é delegada ao AggregationBuffer para evitar
// lock contention em alto volume.
// Seus campos estão alinhados por tamanho em bytes (maior -> menor)
// para otimização de padding do compilador.
type VoteJob struct {
	db     *sql.DB            // 8 bytes
	buffer *AggregationBuffer // 8 bytes
}

// NewVoteJob cria uma nova instância de VoteJob com injeção de dependência do banco e buffer.
func NewVoteJob(db *sql.DB, buffer *AggregationBuffer) *VoteJob {
	return &VoteJob{db: db, buffer: buffer}
}

// Perform insere o voto no banco e incrementa o agregado em memória.
// O agregado será persistido pelo AggregationBuffer de forma batched.
// Exemplo de uso:
//
//	err := job.Perform(ctx, []byte(`{"paredao_id":1,"participant_id":2,"fingerprint_id":"device-1"}`))
func (v *VoteJob) Perform(ctx context.Context, args []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var payload Payload
	if err := json.Unmarshal(args, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal vote payload (offending: %s): %w", string(args), err)
	}

	if payload.ParedaoID == 0 || payload.ParticipantID == 0 {
		return fmt.Errorf("invalid payload (expected non-zero ids, got paredao_id: %d, participant_id: %d)",
			payload.ParedaoID, payload.ParticipantID)
	}

	now := time.Now()

	// Apenas o INSERT de voto — append-only, sem lock de linha hot.
	const insertQuery = `
		INSERT INTO votes (paredao_id, participant_id, fingerprint_id, created_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := v.db.ExecContext(ctx, insertQuery,
		payload.ParedaoID, payload.ParticipantID, payload.Fingerprint, now,
	); err != nil {
		return fmt.Errorf("failed to insert vote into DB: %w", err)
	}

	// Incrementa contador em memória — O(1), sem roundtrip ao banco.
	// O AggregationBuffer fará o flush periódico para vote_aggregations_by_hours.
	v.buffer.Add(payload.ParedaoID, payload.ParticipantID, now.Truncate(time.Hour))

	return nil
}
