package vote

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// VoteJob processa um único voto de forma transacional e atômica.
// Seus campos estão alinhados por tamanho em bytes (maior -> menor)
// para otimização de padding do compilador.
type VoteJob struct {
	db *sql.DB // 8 bytes
}

// NewVoteJob cria uma nova instância de VoteJob com injeção de dependência do banco.
func NewVoteJob(db *sql.DB) *VoteJob {
	return &VoteJob{db: db}
}

// Perform executa a inserção do voto e atualiza a consolidação horária de forma transacional (ACID).
// Exemplo de uso:
//   err := job.Perform(ctx, []byte(`{"paredao_id":1,"participant_id":2,"fingerprint_id":"device-1"}`))
func (v *VoteJob) Perform(ctx context.Context, args []byte) error {
	var payload Payload
	if err := json.Unmarshal(args, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal vote payload (offending: %s): %w", string(args), err)
	}

	if payload.ParedaoID == 0 || payload.ParticipantID == 0 {
		return fmt.Errorf("invalid payload (expected non-zero ids, got paredao_id: %d, participant_id: %d)", payload.ParedaoID, payload.ParticipantID)
	}

	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start vote transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	insertQuery := `
		INSERT INTO votes (paredao_id, participant_id, fingerprint_id, created_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := tx.ExecContext(ctx, insertQuery, payload.ParedaoID, payload.ParticipantID, payload.Fingerprint, now); err != nil {
		return fmt.Errorf("failed to insert vote into DB: %w", err)
	}

	truncatedHour := now.Truncate(time.Hour)
	upsertQuery := `
		INSERT INTO vote_aggregations_by_hours (paredao_id, participant_id, vote_hour, total_votes)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (paredao_id, participant_id, vote_hour)
		DO UPDATE SET total_votes = vote_aggregations_by_hours.total_votes + 1
	`
	if _, err := tx.ExecContext(ctx, upsertQuery, payload.ParedaoID, payload.ParticipantID, truncatedHour); err != nil {
		return fmt.Errorf("failed to upsert vote aggregation: %w", err)
	}

	return tx.Commit()
}
