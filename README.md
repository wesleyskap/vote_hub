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
- **`ingestion/`**: Contém a API em Go (Golang) responsável por receber os votos via HTTP e o worker de processamento que ouve as filas do Redis utilizando a biblioteca [Orkai Runiq](https://github.com/wesleyskap/orkai-runiq) e insere os dados no PostgreSQL.
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

### Pré-requisitos

Antes de começar, certifique-se de ter instalado na sua máquina:

| Ferramenta | Versão mínima | Obrigatório para |
|---|---|---|
| [Docker Desktop](https://www.docker.com/products/docker-desktop/) | 4.x | Ambos os métodos |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | 1.28+ | Kubernetes (método principal) |
| Kubernetes local ativo | — | Kubernetes (Docker Desktop K8s, Minikube ou Kind) |
| [Go](https://go.dev/dl/) | 1.26+ | Testes sem Docker |
| [Ruby](https://www.ruby-lang.org/) | 3.3+ | Testes sem Docker |

---

### 1. Parar serviços conflitantes

Sempre que trocar de método de execução, limpe o ambiente anterior para liberar as portas:

```bash
# Derruba containers do compose
docker compose down

# Remove recursos K8s aplicados anteriormente
kubectl delete -f k8s/

# Encerra port-forwards pendentes no Windows
Stop-Process -Name kubectl -Force
# No Linux / macOS
killall kubectl
```

---

### Método A — Kubernetes (Recomendado para simulação de produção)

> **Requisito:** É necessário ter um cluster Kubernetes local ativo e configurado.
> Opções suportadas: **Docker Desktop (Kubernetes)**, **Minikube** ou **Kind**.

#### Passo 1 — Criar o arquivo de secrets

Copie o arquivo de exemplo e preencha os valores necessários:
```bash
cp k8s/secrets.yaml.example k8s/secrets.yaml
```
> O arquivo `k8s/secrets.yaml` está no `.gitignore` e **não é versionado**.

#### Passo 2 — Buildar as imagens locais

> **Minikube:** Execute `eval $(minikube docker-env)` no mesmo terminal antes de buildar.

```bash
docker build -t bbb-ingestion:local ./ingestion
docker build -t bbb-main-api:local ./backend
docker build -t bbb-frontend:local ./frontend
```

#### Passo 3 — Aplicar todos os manifestos

Se você já tinha o projeto rodando e deseja limpar tudo para subir do zero:
```bash
# Limpa os recursos anteriores
kubectl delete -f k8s/

# Verifica se os pods foram limpos
kubectl get pods
```

Para aplicar os manifestos novamente:
```bash
# Aplica os manifestos
kubectl apply -f k8s/
```

Aguarde todos os pods ficarem prontos (leva cerca de 1-2 minutos):
```bash
kubectl get pods --watch
```
Todos os pods devem estar com status `Running` antes de continuar.

> [!NOTE]
> O script de entrypoint (`docker-entrypoint`) do Rails roda o `db:prepare` automaticamente ao iniciar o pod `main-api`.
> 
> Para verificar se o banco de dados e as tabelas foram criados corretamente:
> ```bash
> # Descubra o nome do pod do postgres
> $POD_NAME = (kubectl get pods -l app=postgres -o jsonpath="{.items[0].metadata.name}")
> 
> # Acesse o psql e liste as tabelas do banco bbb_development
> kubectl exec -it $POD_NAME -- psql -U postgres -d bbb_development -c "\dt"
> ```

#### Passo 4 — Iniciar os túneis de porta (Port-Forward)

```bash
# Windows (PowerShell)
.\start_port_forward.ps1

# Linux / macOS
chmod +x start_port_forward.sh && ./start_port_forward.sh
```

Após executar, os seguintes endereços estarão disponíveis:

| Serviço | Endereço local |
|---|---|
| **Tela de Votação** | http://localhost:3000/ |
| **Parciais ao Vivo** | http://localhost:3000/parciais |
| **Painel Admin** *(interno)* | http://localhost:3000/admin |
| Main API (Rails — Relatórios) | http://localhost:3001 |
| Ingestion API (Go — Votos) | http://localhost:8080 |
| Ingestion API (K6 — Porta NodePort) | http://localhost:30080 |
| Grafana (Painéis de métricas) | http://localhost:3003 |
| Runiq Dashboard (Workers) | http://localhost:8082 |
| Promtail (Service Discovery) | http://localhost:9080/service-discovery |

> **Rotas do frontend:**
> - `/` — tela principal de votação com os cards dos participantes.
> - `/parciais` — resultados ao vivo com polling automático a cada 3s.
> - `/admin` — painel interno de controle com métricas agregadas e histórico por hora. Esta rota **não está linkada publicamente** no header — acesse diretamente pela URL.

---

#### Passo 5 — Rodar o teste de carga com K6 e validação completa

> O K6 roda **dentro do cluster** para evitar gargalos da rede virtual do Docker Desktop.
> Certifique-se de que todos os pods do `ingestion-api` e `ingestion-worker` estão `Running` antes de disparar.

Para rodar os testes e acompanhar os resultados:

```bash
# Deleta o teste anterior se existir
kubectl delete job k6-heavy-test

# Aplica o Job de teste de carga
kubectl apply -f .\k6\k8s-load-test.yaml

# Acompanha a execução do K6 em tempo real
kubectl logs job/k6-heavy-test -c k6 -f
```

Para uma **validação completa**, você também pode simular um erro no Ingestion Worker (para que ele apareça no Grafana Loki) rodando em outro terminal:
```bash
# Windows (PowerShell)
.\simulate_error.ps1
```

Ao final do teste, um relatório customizado será impresso automaticamente:
```
==================================================
        RELATORIO DE PERFORMANCE - K6 (7.5k QPS)
==================================================
  Tempo de Execucao:  180.2s
  Votos Enviados:     1344416
  Vazao Media:        7460.08 req/s

  Taxas de Retorno:
    - Sucesso (202):  100.00% (1344416 votos)
    - Falhas/Outros:  0.00%   (0 requisicoes)

  Tempos de Resposta:
    - Minimo:         0.65ms
    - Mediana (p50):  203.08ms
    - P95:            498.13ms
    - P99 (SLA):      686.42ms
    - Maximo:         5730.00ms
==================================================
```

---

### Método B — Docker Compose (Alternativa simplificada)

> Use este método para desenvolvimento local rápido ou quando não tiver um cluster Kubernetes configurado.
> Este método **não inclui** Grafana, Prometheus, Loki e Promtail.

#### Passo 1 — Subir todos os serviços

```bash
docker compose up --build -d
```

Verifique se os containers estão rodando:
```bash
docker compose ps
```

#### Passo 2 — Preparar o banco de dados

```bash
# Cria o banco, roda as migrations e insere os dados iniciais (paredão + participantes)
docker compose exec main-api bundle exec rails db:prepare db:seed
```

#### Passo 3 — Verificar os endpoints

```bash
# Testa a Ingestion API
curl -X POST http://localhost:8080/api/v1/votes \
  -H "Content-Type: application/json" \
  -d '{"paredao_id":1,"participant_id":1,"fingerprint_id":"test-fp","recaptcha_token":"test-bypass-token"}'

# Testa a Main API
curl http://localhost:3001/api/v1/participants
```

#### Passo 4 — Rodar K6 localmente (opcional)

Com o Docker Compose no ar, o K6 pode ser rodado via Docker apontando para os serviços:
```bash
# Instale o K6: https://grafana.com/docs/k6/latest/set-up/install-k6/
k6 run --env API_URL=http://localhost:8080/api/v1/votes k6/load_test_7k.js
```

---

## Como executar os testes unitários e de integração

### Opção 1 — Sem Docker (requer Go 1.26+ e Ruby 3.3+ instalados localmente)

#### Ingestão (Go)
Testa a API HTTP de votos e o processamento em lote com mocks de banco de dados:
```bash
cd ingestion
go test ./... -v
```

#### Backend (Ruby on Rails)
Requer PostgreSQL acessível. Configure a variável de ambiente antes de rodar:
```bash
cd backend
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/bbb_test"
export RAILS_ENV=test
bundle install
bundle exec rails db:prepare
bundle exec rspec --format documentation
```

#### Frontend (React)
Testa os hooks e o cliente de rede com Vitest (sem dependência de servidor):
```bash
cd frontend
npm install
npm run test
```

---

### Opção 2 — Com Docker Compose (recomendada ao clonar o projeto)

Não requer Go, Ruby ou Node instalados localmente — tudo roda dentro dos containers.

#### Ingestão (Go)
```bash
docker compose run --rm ingestion-api go test ./... -v
```

#### Backend (Ruby on Rails)
```bash
# Sobe apenas o banco de dados
docker compose up -d postgres

# Roda a suíte completa do RSpec em um container isolado
docker compose run --rm \
  -e RAILS_ENV=test \
  -e DATABASE_URL="postgres://postgres:postgres@postgres:5432/bbb_test" \
  main-api bash -c "bundle install --quiet && bundle exec rails db:prepare && bundle exec rspec --format documentation"
```

#### Frontend (React)
```bash
docker compose run --rm frontend npm run test
```



---

## Esteira de CI/CD com GitHub Actions

O repositório possui fluxos de trabalho configurados para automatizar a verificação do código e simular a implantação na AWS:

- **Integração contínua (CI - `.github/workflows/ci.yml`):**
  - Disparado automaticamente em qualquer Pull Request ou push enviado para o branch `master`.
  - Executa testes do Go (`ingestion/`), testes de frontend (`frontend/`) e testes de banco/controllers com PostgreSQL no Rails (`backend/`).
- **Entrega contínua (CD - `.github/workflows/deploy-simulation.yml`):**
  - Disparado ao realizar o merge de novos códigos no branch `master`.
  - Simula o build e envio de imagens Docker para o registro privado (AWS ECR).
  - Executa validações sintáticas e de planejamento do Terraform (`terraform validate`/`plan`).
  - Executa testes de deploy secos (*dry-run*) do Kubernetes (`kubectl apply --dry-run`).

---

### Consulta de logs com Grafana Loki

Para analisar o comportamento da aplicacao e do processamento de votos através do Loki no Grafana, utilize as seguintes queries LogQL no painel Explore:

- **Logs do worker filtrados por severidade (apenas erros):**
  `{app="ingestion-worker"} | json | level="ERROR"`
- **Buscar ciclo completo de um voto usando o Trace ID:**
  `{app=~"ingestion-.*"} | json | trace_id="COLE_O_TRACE_ID_AQUI"`

Filtros rapidos sugeridos:
- `{app="ingestion-worker"}`: logs gerais do worker.
- `{app="ingestion-api"}`: logs gerais da API de ingestao.
- `{app="ingestion-worker"} |= "ERROR"`: busca textual simples por erros no worker.
- `{app="ingestion-worker"} | json | level="WARN"`: apenas logs com severidade de warning no worker.
- `{app=~"ingestion-.*"} | json | level="ERROR"`: logs com severidade de erro de todas as aplicacoes da ingestao.

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
  Rode `.\simulate_error.ps1` para forçar um voto inválido. O erro de Foreign Key Violation gerado no banco deve ser listado nos logs do Loki utilizando o filtro `{app="ingestion-worker"} |= "ERROR"`.
