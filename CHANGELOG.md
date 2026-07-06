# Changelog

Todas as mudancas e evolucoes deste projeto serao documentadas neste arquivo. O formato e baseado no [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/).

---

## [1.0.4] - 2026-07-06

### Corrigido
- Go (CI): atualizado `go-version` de `1.22` para `1.26` no `ci.yml`. O `go.mod` declara `go 1.26.3` (versao instalada localmente), mas o runner instalava apenas o Go 1.22. Com `GOTOOLCHAIN=local`, o runner recusava executar os testes por ser inferior a versao exigida pelo modulo. Removido tambem o `GOTOOLCHAIN=local` que se tornou desnecessario apos a atualizacao da versao.

---

## [1.0.3] - 2026-07-06

### Corrigido
- Go (CI): adicionada a variavel de ambiente `GOTOOLCHAIN=local` no pipeline de CI para impedir que o runner tente baixar automaticamente o toolchain `go1.26.3` declarado no `go.mod`, operacao que falha com `GOSUMDB=off` ativo. O runner passa a usar o toolchain ja instalado sem realizar downloads externos.
- Rails (RSpec): corrigido o teste `GET /api/v1/participants` que retornava `404 Not Found` no CI. O `before(:all)` limpava todos os dados do banco (incluindo o seed), mas o exemplo nao criava um paredao ativo. Adicionados a criacao do paredao e a associacao dos participantes dentro do proprio `it` block, garantindo que `Paredao.find_by(status: active)` nao retorne `nil` em ambientes com banco zerado (como o CI).

---

## [1.0.2] - 2026-07-06

### Modificado
- Reestruturado o fluxo de execucao local nos manuais tecnicos (README.md e README-en.md), definindo o Kubernetes como metodo de deploy principal (cluster obrigatorio) e o Docker Compose como alternativa simplificada de desenvolvimento.
- Movimentado o manifesto do Job de testes de carga do K6 (`k6-load-test.yaml`) para a pasta de testes (`k6/k8s-load-test.yaml`), impedindo execucoes prematuras antes que os servicos do cluster terminem de inicializar.
- Adicionado o callback `handleSummary` em todos os scripts de teste do K6 (`load_test_*.js`) para formatar e imprimir um relatorio de performance limpo e customizado no terminal ao final do teste.
- Atualizado o edital do desafio (`full-stack-challenge.md`) e recompilado o PDF gerado com a nova estrutura de caminhos.

---

## [1.0.1] - 2026-07-06

### Corrigido
- Remocao da diretiva replace local do Go e migracao para o download remoto direto da biblioteca `orkai-runiq/v3` na versao `v3.3.1` do GitHub.
- Adicao de configuracao de bypass de cache de proxy de Go (`GOPROXY=direct` e `GOSUMDB=off`) no pipeline de CI do Actions para evitar falhas por tempo de propagacao.
- Panic de registro duplicado no coletor Prometheus resolvido usando um Registry local por instancia no roteador do Go.
- Falhas de suite de testes do Rails (RSpec) resolvidas com a limpeza automatica do banco de dados de teste, uso do model correto `VoteAggregationsByHour` e correcao das chaves de assercao de respostas da API.

---

## [1.0.0] - 2026-07-06

### Adicionado
- Estrutura base do monorepo unificando os microsservicos do backend (Ruby on Rails 8) e ingestao de votos (Go/Fiber).
- Frontend web desenvolvido em React com hooks sem estado e integrado com fingerprint de hardware e Google reCAPTCHA v3.
- Ingestion Worker em Go consumindo a fila do Redis de forma assincrona e aplicando escrita em lote (Bulk Insert) no PostgreSQL.
- Receitas IaC com Terraform na pasta `terraform/` para provisionamento de VPC, banco RDS PostgreSQL Multi-AZ e cluster AWS EKS.
- Esteira de integracao continua (CI) via GitHub Actions (`ci.yml`) que valida e roda os testes automatizados de Go, Rails e React.
- Esteira de entrega continua (CD) via GitHub Actions (`deploy-simulation.yml`) para simular o build de imagens, validacao do Terraform e deploy seco no Kubernetes EKS.
- Testes unitarios para a API Go de ingestao (`vote_job_test.go`) usando go-sqlmock.
- Testes de requisicao para a API Rails (`participants_spec.rb`) cobrindo listagem de participantes, resultados e estatisticas administrativas.
- Testes unitarios para o hook `useVote` no frontend React usando Vitest e happy-dom.
- Pilha de telemetria e observabilidade local pronta com Prometheus, Grafana Loki e Promtail.
- Dashboards integrados no Grafana para monitoramento de latencias (SLI/SLO), performance do Runiq e dados de negocio.
- Script utilitario de port-forward e script de simulacao de erro (`simulate_error.ps1`) para testar o envio de logs de erro para o Loki.
- Edital do desafio técnico em formato Markdown (`full-stack-challenge.md`) mapeando onde cada requisito foi resolvido no repositorio, acompanhado de script Python (`generate_pdf.py`) para compilacao direta em PDF.
