package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"ingestion/internal/vote"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// FakeQueue simula o comportamento da fila Runiq para testes F.I.R.S.T
type FakeQueue struct {
	EnqueuedJobs int
	ShouldFail   bool
}

func (fq *FakeQueue) Enqueue(ctx context.Context, queueName string, jobType string, payload []byte) error {
	if fq.ShouldFail {
		return errors.New("queue error")
	}
	fq.EnqueuedJobs++
	return nil
}

// FakeRecaptcha simula a validação de bot de forma deterministicamente rápida
type FakeRecaptcha struct {
	ShouldFail bool
	MockResult bool
}

func (fr *FakeRecaptcha) Verify(ctx context.Context, token string, clientIP string) (bool, error) {
	if fr.ShouldFail {
		return false, errors.New("verifier api error")
	}
	return fr.MockResult, nil
}

func TestVoteIngester_Ingest(t *testing.T) {
	t.Run("Voto Valido Enfileira com Sucesso (Fast & Independent)", func(t *testing.T) {
		queue := &FakeQueue{}
		recaptcha := &FakeRecaptcha{MockResult: true}
		ingester := NewVoteIngester(queue, recaptcha)

		app := SetupRouter(ingester)

		payload := vote.Payload{
			Fingerprint:    "device-1",
			RecaptchaToken: "token-valido",
			ParedaoID:      1,
			ParticipantID:  2,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/votes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != fiber.StatusAccepted {
			t.Errorf("expected 202, got %d", resp.StatusCode)
		}
		if queue.EnqueuedJobs != 1 {
			t.Errorf("expected 1 enqueued job, got %d", queue.EnqueuedJobs)
		}
	})

	t.Run("Voto Recusado quando verificação do Google falha (Bot)", func(t *testing.T) {
		queue := &FakeQueue{}
		recaptcha := &FakeRecaptcha{MockResult: false}
		ingester := NewVoteIngester(queue, recaptcha)

		app := SetupRouter(ingester)

		payload := vote.Payload{
			Fingerprint:    "device-1",
			RecaptchaToken: "token-bot",
			ParedaoID:      1,
			ParticipantID:  2,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/votes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
		if queue.EnqueuedJobs != 0 {
			t.Errorf("expected 0 enqueued jobs, got %d", queue.EnqueuedJobs)
		}
	})
}
