# BBB voting system challenge

[**Versão em Português (Portuguese Version)**](./README.md)

Welcome to the Big Brother Brasil voting system repository. This repository unifies the distributed and highly-scalable application designed to receive and process millions of votes.

---

## • Project documentation

### What it is and what it does
The system simulates a reality show eviction night where the audience votes on the available participants to evict them. It ensures:
- Fast edge validation against bots and script manipulation (Rate Limiter, reCAPTCHA v3, and Fingerprinting).
- Asynchronous ingestion that responds to the user in a few milliseconds without locking database traffic.
- Background workers consolidating the vote count.
- Infrastructure telemetry and accumulated results displayed on Grafana dashboards.

### Directory structure and scope
The repository is structured into the following main directories:
- **`ingestion/`**: Contains the Go (Golang) API responsible for receiving votes via HTTP and the background worker consuming Redis queues using the [Orkai Runiq](https://github.com/wesleyskap/orkai-runiq) library to insert data into PostgreSQL.
- **`backend/`**: Contains the administrative API developed in Ruby on Rails 8, responsible for managing the relational database schema and exposing read endpoints for consolidated reports.
- **`frontend/`**: Contains the web application developed in React, structured using stateless custom hooks.
- **`k8s/`**: Manifests for infrastructure resources, services, monitoring (Loki/Grafana/Prometheus), and load-testing scripts configured for Kubernetes.
- **`general/`**: Folder with auxiliary test files, mock data, and obsolete scripts.

---

## • APIs documentation

### 1. Ingestion API (Go - Port 8080)
Focused on high-throughput writing, validation, and fast enqueuing.

- `POST /api/v1/votes`
  - **Function:** Receives and enqueues a vote.
  - **Accepted Payload:**
    ```json
    {
      "paredao_id": 1,
      "participant_id": 2,
      "fingerprint_id": "9a7b9c...",
      "recaptcha_token": "example_token"
    }
    ```
  - **Operation:** Limits traffic to 10 connections/second per IP (Rate Limiter). Validates the reCAPTCHA token against Google's API. Generates a unique Trace ID and enqueues the payload to Redis. Returns `202 Accepted` in less than 5ms.
- `GET /metrics`
  - **Function:** Exposes application metrics in Prometheus format.

### 2. Main API (Rails - Port 3001)
Focused on consistent read queries and administration.

- `GET /api/v1/participants`
  - **Function:** Lists active participants of the eviction round and their metadata.
- `GET /api/v1/results/current`
  - **Function:** Returns the total sum of valid votes for each participant in the active round.
- `GET /admin/v1/stats`
  - **Function:** Endpoint consumed by the administrative dashboard to fetch aggregate stats and throughput rate (QPS).

---

## • Architecture documentation

### Asynchronous Bulk Insert
Inserting each vote individually into PostgreSQL would cause disk contention and exhaust connection pools under heavy load (e.g., 7,500 RPS). The Go Worker uses a thread-safe `AggregationBuffer`. Data insertion only occurs periodically or when a batch size limit is met, grouping thousands of transactions into a single `INSERT INTO ... ON CONFLICT DO UPDATE` query.

### Distributed Tracing (Trace ID)
Upon receiving a vote in the Ingestion API, a unique tracking identifier (`Trace ID`) is attached to the payload. This ID travels through the Redis queue and is extracted by the Ingestion Worker. Logs are structured in JSON format and collected by Promtail. In case of a database write failure (e.g., Foreign Key violation), matching the Trace ID on Grafana Loki allows developers to trace the exact workflow path of that vote.

### Stateless Custom Hooks
Vote submission logic, reCAPTCHA flows, and canvas fingerprint calculations are decoupled from visual components and encapsulated in `frontend/src/hooks/useVote.js`. This DOM-agnostic pattern simplifies future porting to mobile technologies (such as React Native) with minimal code refactoring.

---

## • Documentation on how to spin up a copy of this environment locally

### 1. Stop conflicting services
Ensure local host ports are free by cleaning previous executions:
```bash
# Tear down running compose services
docker-compose down

# Delete applied K8s resources
kubectl delete -f k8s/

# Terminate pending kubectl port-forward processes on Windows
Stop-Process -Name kubectl -Force
# Or on Linux / macOS
killall kubectl
```

### 2. Running with Docker Compose
Recommended for simple local testing.
```bash
# Spin up db, APIs, frontend, and queues
docker-compose up --build -d

# Run migrations and seed data
docker compose exec main-api bundle exec rails db:prepare db:seed
```

### 3. Running with Kubernetes
Recommended for production simulations and stress tests.

#### Compile local Docker images
Generate local builds so the Kubernetes cluster can load them:
```bash
docker build -t bbb-ingestion:local ./ingestion
docker build -t bbb-main-api:local ./backend
docker build -t bbb-frontend:local ./frontend
```
*(Note: If using Minikube, execute `eval $(minikube docker-env)` in the terminal first).*

#### Apply manifests
```bash
kubectl apply -f k8s/
```

#### Enable Port-Forward tunnels
Expose the Kubernetes internal network ports to your localhost:
- **Windows (PowerShell):** `.\start_port_forward.ps1`
- **Linux / macOS (Bash):** `chmod +x start_port_forward.sh && ./start_port_forward.sh`

---

### Running stress tests
To run heavy concurrency simulations of up to 7,500 requests per second, the K6 load test must run inside the cluster to avoid virtual network host adapter bottlenecks:
```bash
# Apply the load test job
kubectl apply -f k8s/k6-load-test.yaml

# Monitor live reports and console outputs
kubectl logs -f job/k6-heavy-test -c k6
```

---

### Troubleshooting

- **Loki offline or "Failed to load log volume":**
  Heavy K6 traffic logs can crash Loki due to out-of-memory errors (OOMKilled). The memory limit has been increased from `512Mi` to `2Gi` in `k8s/loki.yaml`. If this persists, run `kubectl rollout restart deployment/loki`.
- **Promtail with empty target list (Service Discovery empty):**
  Ensure the `HOSTNAME` environment variable is mapped in the Promtail daemon container based on the node's metadata (`spec.nodeName`) so it can access Docker log folders.
- **Local ports already in use:**
  If you encounter port binding errors during start-up, terminate conflicting processes or force close all kubectl commands:
  ```powershell
  Stop-Process -Name kubectl -Force
  ```
- **Database integrity errors when testing failures:**
  Run `.\general\simulate_error.ps1` to force an invalid vote. The resulting Foreign Key violation will be visible in the Loki dashboard under the filter `{app="ingestion-worker"} |= "ERROR"`.
