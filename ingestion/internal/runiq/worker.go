package runiq

import (
	"context"

	"github.com/wesleyskap/orkai-runiq/v3/queue"
)

// Job define o comportamento que toda tarefa de background do sistema deve possuir.
type Job interface {
	Perform(ctx context.Context, args []byte) error
}

// WorkerPool encapsula o pool de workers concorrentes da fila do Runiq (Wrapper Pattern).
// Seus campos estão alinhados por tamanho em bytes (maior -> menor)
// para otimização de padding do compilador.
type WorkerPool struct {
	pool *queue.WorkerPool // 8 bytes
}

// NewWorkerPool cria um novo pool de workers delegando internamente para a biblioteca Runiq.
// Exemplo de uso:
//   workerPool := runiq.NewWorkerPool(storage, 10)
func NewWorkerPool(storage queue.WorkerPoolStorage, concurrency int, opts ...queue.WorkerOption) *WorkerPool {
	return &WorkerPool{
		pool: queue.NewWorkerPool(storage, concurrency, opts...),
	}
}

// Register associa um tipo de job a um handler implementando a interface local Job.
// Exemplo de uso:
//   workerPool.Register("process_vote", myJob)
func (w *WorkerPool) Register(name string, job Job) {
	w.pool.Register(name, &jobAdapter{job: job})
}

// Start inicia o consumo concorrente e bloqueante de uma fila específica.
// Exemplo de uso:
//   err := workerPool.Start(ctx, "votes_queue")
func (w *WorkerPool) Start(ctx context.Context, queueName string) error {
	return w.pool.Start(ctx, queueName)
}

type jobAdapter struct {
	job Job // 16 bytes (interface structure: type descriptor pointer + data pointer)
}

func (ja *jobAdapter) Perform(ctx context.Context, args []byte) error {
	return ja.job.Perform(ctx, args)
}
