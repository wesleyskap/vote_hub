# EvolucaoFutura da Infraestrutura (Escala BBB)

Este documento descreve as melhorias estrategicas recomendadas para suportar picos de votacao acima de 100.000 requisicoes por segundo na AWS.

---

## EscalonamentoBaseadoFila com KEDA

Hoje o Kubernetes escala os Pods medindo CPU e memoria (HPA). Em processos de fila (como o Worker Pool lendo do Redis), isso nao e o ideal porque a fila pode crescer muito rapido sem que a CPU suba no mesmo ritmo.

### Solucao
Instalar o **KEDA** (Kubernetes Event-driven Autoscaling) e monitorar o tamanho da fila:
- **Metrica:** Quantidade de itens pendentes na fila (`LLEN votes_queue`).
- **Acao:** Se a fila acumular mais de 5.000 votos, o KEDA escala o Worker Pool de 3 para ate 50 Pods em segundos, limpando o gargalo.

---

## Camada de cache e leitura segura

Evitar que consultas de leitura do painel de resultados (gerido pela API Rails) travem o banco de dados principal de gravacao de votos.

### Solucao
- **CacheAside:** Ler o total de votos direto do Redis. Se nao existir, busca no RDS e salva no Redis por alguns segundos.
- **ReplicaLeitura:** Configurar uma replica de leitura no RDS (Read Replica). O Rails deve enviar consultas pesadas (como graficos e dashboards) apenas para essa replica, deixando o banco principal livre para gravar votos.

---

## Protecao Edge e mitigacao DDoS

Mitigar ataques de robos e bots na borda (Edge) antes que a carga atinja nossos servidores internos na VPC.

### Solucao
- **CloudFront com WAF:** Adicionar o AWS WAF na frente da Ingestion API.
- **Regras:**
  - Limite de requisicoes por IP (ex: maximo de 20 votos por minuto por IP residencial).
  - Bloqueio de paises fora do publico alvo.
  - Bloqueio de proxies e IPs suspeitos.

---

## Particionamento dados e coldStorage

Conforme o volume de votos cresce, a tabela principal do PostgreSQL fica lenta e cara.

### Solucao
- **ParticionamentoMensal:** Particionar a tabela de votos por dia ou hora no PostgreSQL.
- **ColdStorage:** Rodar um script ao fim do paredao que exporta os votos antigos em formato Parquet para o Amazon S3 (armazenamento barato) e apaga os dados antigos do banco ativo.
