package runiq

import (
	"context"

	"github.com/wesleyskap/orkai-runiq/v3/queue"
)

// Client encapsula o client da biblioteca de mensageria de terceiros (Wrapper Pattern).
// Seus campos estão alinhados por tamanho em bytes (maior -> menor)
// para otimização de padding do compilador.
type Client struct {
	qClient *queue.Client // 8 bytes
}

// NewClient cria um novo wrapper para o cliente da fila do Runiq.
// Exemplo de uso:
//   client := runiq.NewClient(storage)
func NewClient(storage queue.ClientStorage) *Client {
	return &Client{
		qClient: queue.NewClient(storage),
	}
}

// Enqueue enfileira uma tarefa delegando para o cliente interno da biblioteca.
// Exemplo de uso:
//   err := client.Enqueue(ctx, "default", "job_type", []byte(`{}`))
func (c *Client) Enqueue(ctx context.Context, queueName, jobType string, payload []byte) error {
	return c.qClient.Enqueue(ctx, queueName, jobType, payload)
}
