package api

import (
	"context"
	"encoding/json"
	"ingestion/internal/vote"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

// JobEnqueuer define a interface mínima requerida pelo Ingester (Consumer-side Interface)
type JobEnqueuer interface {
	Enqueue(ctx context.Context, queueName string, jobType string, payload []byte) error
}

// TokenVerifier define a interface mínima para validação de captcha
type TokenVerifier interface {
	Verify(ctx context.Context, token string, clientIP string) (bool, error)
}

// VoteIngester orquestra a recepção rápida e enfileiramento dos votos
type VoteIngester struct {
	enqueuer JobEnqueuer   // 8 bytes
	verifier TokenVerifier // 8 bytes
}

// NewVoteIngester construtor por injeção direta sem variáveis globais
func NewVoteIngester(enqueuer JobEnqueuer, verifier TokenVerifier) *VoteIngester {
	return &VoteIngester{
		enqueuer: enqueuer,
		verifier: verifier,
	}
}

// Ingest processa a requisição, valida se é bot, e enfileira no Runiq
func (h *VoteIngester) Ingest(c *fiber.Ctx) error {
	var payload vote.Payload
	if err := c.BodyParser(&payload); err != nil {
		slog.Error("invalid payload format received", "err", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	// Early Returns: Indentidade máxima de 2 níveis
	if payload.ParedaoID == 0 || payload.ParticipantID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ids are required"})
	}

	ok, err := h.verifier.Verify(c.Context(), payload.RecaptchaToken, c.IP())
	if err != nil || !ok {
		slog.Warn("verification failed or suspect bot", "ip", c.IP(), "err", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bot protection triggered"})
	}

	rawPayload, _ := json.Marshal(payload)
	if err := h.enqueuer.Enqueue(c.Context(), "votes_queue", "process_vote", rawPayload); err != nil {
		slog.Error("failed to queue", "err", err.Error())
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "service temporary busy"})
	}

	return c.SendStatus(fiber.StatusAccepted)
}
