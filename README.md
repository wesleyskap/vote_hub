# Desafio técnico - Sistema de votação BBB

[English Version (Versão em Inglês)](./README-en.md)

Bem-vindo ao repositório do sistema de votação do Big Brother Brasil. Este repositório unifica a aplicação distribuída e de alta escalabilidade desenhada para receber e processar milhões de votos.

---

## Documentação do projeto

Para detalhes da arquitetura de software, diagramas C4 e decisões de design, consulte o arquivo [Arquitetura.md](/Arquitetura.md).

### O que é e o que faz
O sistema simula um paredão de reality show onde o público vota nos participantes disponíveis para eliminação. Ele garante:
- Validação rápida na borda contra bots e cliques repetidos (Rate limiter, reCAPTCHA v3 e fingerprint).
- Ingestão assíncrona que responde ao usuário em poucos milissegundos sem travar o tráfego no banco de dados.
- Processamento em segundo plano (background workers) consolidando a contagem de votos.
- Telemetria de infraestrutura e resultados acumulados exibidos em painéis do Grafana.

### Divisão de pastas e escopo
O repositório está estruturado nos seguintes diretórios principais:
- **`ingestion/`**: Contém a API em Go (Golang) responsável por receber os votos via HTTP e o worker de processamento que ouve as filas do Redis e insere dados no PostgreSQL.
- **`backend/`**: Contém a API administrativa desenvolvida em Ruby on Rails 8, responsável por manter o schema do banco de dados relacional e expor as rotas de leitura dos relatórios consolidados.
- **`frontend/`**: Contém a aplicação web desenvolvida em React e estruturada por meio de hooks sem estado (stateless) reutilizáveis.
- **`k8s/`**: Manifesto de recursos, serviços, monitoramento (Loki/Grafana/Prometheus) e testes de estresse estruturados para o Kubernetes.
- **`terraform/`**: Configuração de infraestrutura como código (IaC) para provisionar os recursos na nuvem AWS (VPC, EKS, RDS PostgreSQL Multi-AZ).
- **`general/`**: Pasta com arquivos auxiliares de teste, mocks de dados e scripts obsoletos.

---

## Documentação das APIs

### Ingestion API (Go - Porta 8080)
Focada em escrita, validação inicial e enfileiramento rápido.

- `POST /api/v1/votes`
  - **Função:** Recebe e enfileira o voto.
  - **Payload aceito:**
    ```json
    {
      "paredao_id": 1,
      "participant_id": 2,
      "fingerprint_id": "9a7b9c...",
      "recaptcha_token": "token_exemplo"
    }
    ```
  - **Funcionamento:** Limita a 10 conexões/segundo por IP (Rate Limiter). Valida o token reCAPTCHA contra a API do Google. Gera um Trace ID único e empilha a tarefa no Redis. Retorna status `202 Accepted` em até 5ms.
- `GET /metrics`
  - **Função:** Fornece as métricas da aplicação em formato legível pelo Prometheus.

### Main API (Rails - Porta 3001)
Focada em leitura consistente e administração.

- `GET /api/v1/participants`
  - **Função:** Lista os participantes ativos do paredão e seus metadados (como URLs de imagens).
- `GET /api/v1/results/current`
  - **Função:** Retorna a somatória acumulada de votos válidos de cada participante do paredão ativo.
- `GET /admin/v1/stats`
  - **Função:** Rota consumida pelo painel administrativo para obter estatísticas agregadas e taxa de vazão (QPS).

---

## Arquitetura geral

### Inserção assíncrona em lote (Bulk Insert)
Inserir individualmente cada voto no PostgreSQL geraria contenção de disco e travamento de conexões sob cargas intensas (ex: 7.500 RPS). O Worker em Go utiliza um `AggregationBuffer` thread-safe. A inserção no banco de dados só ocorre de forma acumulada a cada intervalo de tempo ou lote preenchido, convertendo milhares de transações em um único comando `INSERT INTO ... ON CONFLICT DO UPDATE`.

### Rastreabilidade distribuída (Trace ID)
Ao receber um voto na Ingestion API, um identificador único de rastreamento (`Trace ID`) é atrelado ao payload. Esse ID transita pela fila do Redis e é extraído pelo Ingestion Worker. Os logs são estruturados em formato JSON e capturados pelo Promtail. Em caso de falha silenciosa de integridade (ex: Foreign Key Violation no banco), o Trace ID nos logs do Grafana Loki permite cruzar a origem exata do payload com o respectivo erro gerado.

### Desacoplamento da interface (Hooks stateless)
A lógica de submissão do voto, controle do reCAPTCHA e obtenção do fingerprint no React estão isoladas no arquivo `frontend/src/hooks/useVote.js`. Esse padrão agnóstico do DOM simplifica a portabilidade futura para tecnologias mobile (como React Native) com pouca refatoração.

---

## Como subir uma cópia deste ambiente localmente

### Parar serviços conflitantes
Certifique-se de liberar as portas do seu host local limpando execuções anteriores:
```bash
# Derruba containers ativos do compose
docker-compose down

# Remove recursos aplicados do K8s
kubectl delete -f k8s/

# Encerra processos do kubectl pendentes no Windows
Stop-Process -Name kubectl -Force
# Ou no Linux / macOS
killall kubectl
```

### Executando com Docker compose
Recomendado para testes simples locais.
```bash
# Sobe banco, APIs, frontend e filas
docker-compose up --build -d

# Prepara as tabelas e adiciona massa de testes
docker compose exec main-api bundle exec rails db:prepare db:seed
```

### Executando com Kubernetes
Recomendado para simulação de produção e testes de estresse.

#### Compilar as imagens
Gere os binários sob a tag local para que o cluster K8s possa consumi-los:
```bash
docker build -t bbb-ingestion:local ./ingestion
docker build -t bbb-main-api:local ./backend
docker build -t bbb-frontend:local ./frontend
```
*(Nota: Se usar Minikube, execute primeiro `eval $(minikube docker-env)` no mesmo terminal).*

#### Aplicar manifestos
```bash
kubectl apply -f k8s/
```

#### Habilitar túneis de porta (Port-Forward)
Para expor os endpoints das redes internas do Kubernetes para o seu computador:
- **No Windows (PowerShell):** `.\start_port_forward.ps1`
- **No Linux / macOS (Bash):** `chmod +x start_port_forward.sh && ./start_port_forward.sh`

---

### Executando testes de estresse
Para rodar simulações de carga concorrente pesada de até 7.500 requisições por segundo, o teste K6 deve rodar dentro do próprio cluster para evitar gargalos na interface virtual de rede do host local:
```bash
# Dispara o job de teste de estresse
kubectl apply -f k8s/k6-load-test.yaml

# Visualiza os relatórios gerados em tempo real
kubectl logs -f job/k6-heavy-test -c k6
```

---

## Como executar os testes unitários e de integração

Você pode rodar os testes automatizados em cada ponta do projeto usando os comandos abaixo:

### Ingestão (Go)
Testa a API HTTP de votos e o processamento em lote do banco de dados (usando mocks/sqlmock).
```bash
cd ingestion
go test ./... -v
```

### Backend (Ruby on Rails)
Testa as regras dos modelos de participantes e os controllers de API REST com RSpec.
```bash
# Com o ambiente docker-compose ativo
docker compose exec main-api bundle exec rspec
```

### Frontend (React)
Testa as abstrações do cliente de rede e o estado lógico do hook useVote com Vitest.
```bash
cd frontend
npm run test
```

---

### Diagnóstico de problemas (Troubleshooting)

- **Loki offline ou "Failed to load log volume":**
  A alta volumetria de logs gerada pelo K6 pode derrubar o Loki por falta de memória (OOMKilled). Aumentamos o limite padrão de memória de `512Mi` para `2Gi` no manifesto `k8s/loki.yaml`. Caso ocorra, aplique `kubectl rollout restart deployment/loki`.
- **Promtail com lista de alvos vazia (Service Discovery vazio):**
  Certifique-se de que a variável de ambiente `HOSTNAME` está sendo alimentada no container do Promtail com base no metadado do nó (`spec.nodeName`) para que ele consiga ler as pastas de logs locais do Docker.
- **Portas locais em uso:**
  Se receber erros de porta ocupada (bind) ao subir o túnel do K8s, encerre os processos ocupando as portas locais ou force o fechamento do kubectl:
  ```powershell
  Stop-Process -Name kubectl -Force
  ```
- **Erro de integridade em banco ao testar falhas:**
  Rode `.\general\simulate_error.ps1` para forçar um voto inválido. O erro de Foreign Key Violation gerado no banco deve ser listado nos logs do Loki utilizando o filtro `{app="ingestion-worker"} |= "ERROR"`.
