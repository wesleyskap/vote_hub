package vote

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestVoteJob_Perform(t *testing.T) {
	t.Run("deve inserir voto valido no banco e adicionar no buffer", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("falha ao criar mock sql: %v", err)
		}
		defer db.Close()

		// Configura expectativa do banco de dados
		mock.ExpectExec("INSERT INTO votes").
			WithArgs(int64(1), int64(2), "fingerprint-123", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		buffer := NewAggregationBuffer(db, 5*time.Second, 100)
		job := NewVoteJob(db, buffer)

		payload := Payload{
			ParedaoID:     1,
			ParticipantID: 2,
			Fingerprint:   "fingerprint-123",
			TraceID:       "trace-uuid-1",
		}
		args, _ := json.Marshal(payload)

		ctx := context.Background()
		err = job.Perform(ctx, args)
		if err != nil {
			t.Errorf("esperava sucesso, obteve erro: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas do banco nao atendidas: %v", err)
		}

		// Valida se foi incrementado no buffer de agregacao
		buffer.mu.Lock()
		defer buffer.mu.Unlock()
		if len(buffer.counts) != 1 {
			t.Errorf("esperava 1 item no buffer, obteve %d", len(buffer.counts))
		}
	})

	t.Run("deve retornar erro se o payload for malformado", func(t *testing.T) {
		db, _, _ := sqlmock.New()
		defer db.Close()

		buffer := NewAggregationBuffer(db, 5*time.Second, 100)
		job := NewVoteJob(db, buffer)

		ctx := context.Background()
		err := job.Perform(ctx, []byte(`{invalid-json`))
		if err == nil {
			t.Error("esperava erro de payload malformado, obteve nil")
		}
	})

	t.Run("deve retornar erro se os IDs obrigatorios estiverem ausentes", func(t *testing.T) {
		db, _, _ := sqlmock.New()
		defer db.Close()

		buffer := NewAggregationBuffer(db, 5*time.Second, 100)
		job := NewVoteJob(db, buffer)

		payload := Payload{
			Fingerprint: "fingerprint-123",
			ParedaoID:   0, // ID zerado
		}
		args, _ := json.Marshal(payload)

		ctx := context.Background()
		err := job.Perform(ctx, args)
		if err == nil {
			t.Error("esperava erro por IDs zerados, obteve nil")
		}
	})

	t.Run("deve retornar erro se a gravacao no banco falhar", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()

		mock.ExpectExec("INSERT INTO votes").
			WillReturnError(errors.New("db connection failure"))

		buffer := NewAggregationBuffer(db, 5*time.Second, 100)
		job := NewVoteJob(db, buffer)

		payload := Payload{
			ParedaoID:     1,
			ParticipantID: 2,
			Fingerprint:   "fingerprint-123",
		}
		args, _ := json.Marshal(payload)

		ctx := context.Background()
		err := job.Perform(ctx, args)
		if err == nil {
			t.Error("esperava erro por falha de banco, obteve nil")
		}
	})
}
