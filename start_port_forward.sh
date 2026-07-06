#!/bin/bash

# Script Bash para iniciar e gerenciar todos os port-forwards do Kubernetes no Linux / macOS.
# Como executar: chmod +x start_port_forward.sh && ./start_port_forward.sh

echo "Limpando port-forwards antigos do kubectl..."
killall kubectl 2>/dev/null || true

echo "Iniciando novos túneis de Port-Forward para o Kubernetes..."

# Inicia os redirecionamentos em background
kubectl port-forward svc/frontend 3000:3000 >/dev/null 2>&1 &
kubectl port-forward svc/main-api 3001:3001 >/dev/null 2>&1 &
kubectl port-forward svc/ingestion-api 8080:8080 >/dev/null 2>&1 &
kubectl port-forward svc/ingestion-api 30080:30080 >/dev/null 2>&1 &
kubectl port-forward svc/grafana 3003:3003 >/dev/null 2>&1 &
kubectl port-forward svc/ingestion-worker 8082:8082 >/dev/null 2>&1 &
kubectl port-forward svc/promtail 9080:9080 >/dev/null 2>&1 &

echo "---------------------------------------------------------"
echo "Todos os Port-Forwards foram iniciados com sucesso!"
echo "Endereços para acesso local:"
echo "  - Interface de Votação (Frontend): http://localhost:3000"
echo "  - API Principal (Rails Reports):   http://localhost:3001"
echo "  - Ingestion API (Go Endpoint):     http://localhost:8080"
echo "  - Ingestion API (K6 Target):       http://localhost:30080"
echo "  - Grafana (Painéis de Métricas):   http://localhost:3003"
echo "  - Runiq Dashboard (Workers UI):    http://localhost:8082"
echo "  - Promtail (Service Discovery):    http://localhost:9080/service-discovery"
echo "---------------------------------------------------------"
echo "Testes:"
echo "Para simular um erro no Ingestion Worker (aparecerá no Grafana Loki), execute:"
echo "pwsh ./simulate_error.ps1"
echo "---------------------------------------------------------"
echo "Para encerrar todos os port-forwards a qualquer momento, execute:"
echo "killall kubectl"
