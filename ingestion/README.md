# Ingestion API & worker

Este diretório contém o core de absorção de tráfego do projeto: a API de ingestão de votos e o Ingestion Worker, ambos desenvolvidos em Go (Golang) com foco extremo em performance.

## Mapeamento de rotas (Ingestion API)

A Ingestion API é a única interface de gravação exposta para o mundo externo.

- `POST /api/v1/votes`
  - **Função:** Recebe o payload do voto contendo `participant_id`, `paredao_id`, `fingerprint_id` e o `recaptcha_token`.
  - **Mecanismos críticos:** 
    1. **Rate Limiter de Memória:** O Fiber intercepta a chamada limitando a 10 requisições por segundo por IP. Se exceder, retorna `429 Too Many Requests`.
    2. **Integração reCAPTCHA:** O token é despachado via HTTP POST para a API do Google para aferir a validade do usuário.
    3. **Enfileiramento (Enqueue):** Se o voto for íntegro, um `Trace ID` único é gerado (para rastreabilidade no Grafana Loki) e o payload é submetido à fila do Redis usando a biblioteca [Orkai Runiq](https://github.com/wesleyskap/orkai-runiq). O cliente recebe uma resposta HTTP `202 Accepted` em menos de ~5ms, isolando-o completamente de latências do banco de dados.

- `GET /metrics`
  - **Função:** Rota não-autenticada consumida exclusivamente pela infraestrutura (Prometheus) para coleta do SLI de telemetria da aplicação Go.

## Decisões arquiteturais

1. **Golang para alta concorrência:** Go garante o gerenciamento de milhares de conexões simultâneas (goroutines) na borda de ingestão com baixo custo de memória, superando frameworks tradicionais baseados em threads.
2. **Desacoplamento via mensageria (Redis):** A API não escreve diretamente no banco de dados. A validação do voto ocorre em milissegundos, com o payload empilhado na memória do Redis, liberando o cliente imediatamente e evitando gargalos de I/O em picos de 1.000 a 7.500 votos por segundo.
3. **Padrão de worker e bulk insert:** O worker consome a fila do Redis e agrega os votos em memória (Aggregation buffer). Os agrupamentos são inseridos em lote no banco de dados (Bulk Inserts), reduzindo milhares de transações para apenas uma.
   - *Como funciona no código (`ingestion/internal/vote/aggregation_buffer.go`):* O Worker possui um `map[aggKey]int64` protegido por um `sync.Mutex`. Ao invés de fazer um `INSERT` a cada voto do Redis, ele faz `buffer.counts[key]++`. Um *Ticker* em background consolida a memória e escreve de uma vez usando `INSERT INTO ... ON CONFLICT DO UPDATE`. O banco de dados recebe 1 transação por segundo em vez de milhares.
4. **Motor [Orkai Runiq](https://github.com/wesleyskap/orkai-runiq):** Gerencia a pool de workers, fornecendo retry automático nativo e roteamento para uma Dead Letter Queue (DLQ) em falhas permanentes (ex: violação de Foreign Key).

## Estrutura de pastas

```text
ingestion/
├── cmd/
│   ├── api/          # Ponto de entrada da API HTTP de ingestão. Inicializa o roteador Fiber.
│   └── worker/       # Ponto de entrada do Worker [Orkai Runiq](https://github.com/wesleyskap/orkai-runiq). Conecta no Redis e banco.
├── internal/
│   ├── api/          # Lógica de rotas e validações do HTTP (handlers).
│   └── vote/         # Core business: Definição dos payloads, filas, Aggregation Buffer e persistência.
├── go.mod / go.sum   # Gerenciamento de dependências.
└── Dockerfile        # Compilação otimizada em dois estágios (multi-stage build) para menor tamanho.
```
## Tecnologias

- **Go** 1.26
- **Fiber** v2 (framework web de alta performance)
- **Redis Go Client** v9 (interação assíncrona de fila)
- **[Orkai Runiq](https://github.com/wesleyskap/orkai-runiq)** v3 (gerenciador de background jobs e DLQ)
- **PostgreSQL Driver** (`github.com/lib/pq` para bulk inserts)
- 
## Componentes

- **Ingestion API**: Serviço de borda de altíssima vazão. Responsável por receber as requisições HTTP de votos, realizar validações em tempo real (como IDs nulos e tokens de bot) e encaminhar as mensagens de forma assíncrona para a fila do Redis utilizando o motor [Orkai Runiq](https://github.com/wesleyskap/orkai-runiq). Essa API não faz escritas diretas no banco relacional.
- **Ingestion worker**: Um processo operário distribuído e auto-escalável que ouve o Redis. Ele agrega a contagem de votos diretamente em um buffer de memória e os insere no banco de dados em lote (Bulk Insert) em um fluxo controlado.

## Mapeamento de rotas (Ingestion API)

- `POST /api/v1/votes`
  - **Função:** Recebe o payload do voto contendo `participant_id`, `paredao_id`, `fingerprint_id` e o `recaptcha_token`.
  - **Mecanismos críticos:**
    1. **Rate Limiter de Memória:** O Fiber intercepta a chamada limitando a 10 requisições por segundo por IP. Se exceder, retorna `429 Too Many Requests`.
    2. **Integração reCAPTCHA:** O token é despachado via HTTP POST para a API do Google para aferir a validade do usuário.
    3. **Enfileiramento (Enqueue):** Se o voto for íntegro, um `Trace ID` único é gerado (para rastreabilidade no Loki) e o payload é submetido à fila do Redis usando a biblioteca [Orkai Runiq](https://github.com/wesleyskap/orkai-runiq). O cliente recebe uma resposta HTTP `202 Accepted` em menos de ~5ms.

- `GET /metrics`
  - **Função:** Rota consumida pelo Prometheus para coleta de telemetria da aplicação Go.

## Rastreabilidade e prevenção a falhas invisíveis

Todas as aplicações Go neste módulo implementam logging estruturado em formato JSON e injetam automaticamente um "Trace ID" por meio do contexto (`context.Context`), garantindo que um voto individual possa ser auditado e localizado nos dashboards do Loki/Grafana desde a entrada da HTTP API até o momento em que atinge o banco de dados.

- Quando a API aceita o voto, ela capta (ou cria) o header `X-Trace-Id` e embute isso na raiz do payload da mensagem.
- Quando o Worker processa a mensagem, ele lê o `trace_id` e o anexa ao logger estruturado (`slog.With("trace_id", payload.TraceID)`).
- O Kubernetes captura a saída padrão via **Promtail** e despacha para o **Grafana Loki**. Se um dos votos estourar um erro complexo de integridade no banco (ex: Foreign Key Violation), você pode colar o `Trace ID` no Grafana e ver a linha do tempo exata: desde a entrada no Fiber até a morte da rotina no Worker. Sem esse padrão, debugar hiperconcorrência seria impossível.

## Comandos úteis

```bash
# Executar a API de ingestão localmente (fora do Docker)
go run cmd/api/main.go

# Executar o Worker de consumo localmente (fora do Docker)
go run cmd/worker/main.go

# Rodar todos os testes unitários e de integração do Go
go test ./... -v
```

## Troubleshooting (Resolução de problemas)

- **Erro "Unable to initialize redis storage":**
  Certifique-se de que o container do Redis está rodando e acessível na porta configurada. Se estiver rodando o Go localmente direto na máquina física, certifique-se de alterar as configurações de endereço no arquivo `cmd/api/main.go` e `cmd/worker/main.go` ou configure os port-forwards se estiver testando contra o cluster Kubernetes.

- **Conexão com PostgreSQL caindo no Worker:**
  O worker necessita de uma conexão estável com o banco para realizar o flush periódico. Em caso de quedas ou lentidão nas transações, aumente o tempo de timeout ou verifique se o pool de conexões do PostgreSQL (`SetMaxIdleConns`/`SetMaxOpenConns`) do Go não atingiu o limite configurado no arquivo `cmd/worker/main.go`.
