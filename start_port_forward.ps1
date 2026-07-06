# Script PowerShell para iniciar e gerenciar todos os port-forwards do Kubernetes localmente.
# Como executar: .\start_port_forward.ps1

# 1. Limpa port-forwards travados na máquina
Write-Host "Limpando port-forwards antigos do kubectl..." -ForegroundColor Yellow
Stop-Process -Name kubectl -Force -ErrorAction SilentlyContinue

Write-Host "Iniciando novos túneis de Port-Forward para o Kubernetes..." -ForegroundColor Green

# 2. Inicia os redirecionamentos em processos ocultos (background) no Windows
# Frontend (React Web) - Mapeia localhost:3000 para a porta 3000 do Service
Start-Process kubectl -ArgumentList "port-forward svc/frontend 3000:3000" -WindowStyle Hidden

# Main API (Ruby on Rails) - Mapeia localhost:3001 para a porta 3001 do Service
Start-Process kubectl -ArgumentList "port-forward svc/main-api 3001:3001" -WindowStyle Hidden

# Ingestion API (Go - Votos padrão) - Mapeia localhost:8080 para a porta 8080 do Service
Start-Process kubectl -ArgumentList "port-forward svc/ingestion-api 8080:8080" -WindowStyle Hidden

# Ingestion API (Go - K6 script) - Mapeia localhost:30080 para a porta 30080 do Service
Start-Process kubectl -ArgumentList "port-forward svc/ingestion-api 30080:30080" -WindowStyle Hidden

# Grafana (Métricas) - Mapeia localhost:3003 para a porta 3003 do Service
Start-Process kubectl -ArgumentList "port-forward svc/grafana 3003:3003" -WindowStyle Hidden

# Runiq Dashboard (Workers) - Mapeia localhost:8082 para a porta 8082 do Service
Start-Process kubectl -ArgumentList "port-forward svc/ingestion-worker 8082:8082" -WindowStyle Hidden

# Promtail (Service Discovery / SD HTML) - Mapeia localhost:9080 para a porta 9080 do Service
Start-Process kubectl -ArgumentList "port-forward svc/promtail 9080:9080" -WindowStyle Hidden

# 3. Informações de Acesso
Write-Host "---------------------------------------------------------" -ForegroundColor Gray
Write-Host "Todos os Port-Forwards foram iniciados com sucesso!" -ForegroundColor Green
Write-Host "Endereços para acesso local:" -ForegroundColor Cyan
Write-Host "  - Interface de Votação (Frontend): http://localhost:3000" -ForegroundColor Cyan
Write-Host "  - API Principal (Rails Reports):   http://localhost:3001" -ForegroundColor Cyan
Write-Host "  - Ingestion API (Go Endpoint):     http://localhost:8080" -ForegroundColor Cyan
Write-Host "  - Ingestion API (K6 Target):       http://localhost:30080" -ForegroundColor Cyan
Write-Host "  - Grafana (Painéis de Métricas):   http://localhost:3003" -ForegroundColor Cyan
Write-Host "  - Runiq Dashboard (Workers UI):    http://localhost:8082" -ForegroundColor Cyan
Write-Host "  - Promtail (Service Discovery):    http://localhost:9080/service-discovery" -ForegroundColor Cyan
Write-Host "---------------------------------------------------------" -ForegroundColor Gray
Write-Host "Testes:" -ForegroundColor Magenta
Write-Host "Para simular um erro no Ingestion Worker (aparecerá no Grafana Loki), execute:" -ForegroundColor Yellow
Write-Host ".\general\simulate_error.ps1" -ForegroundColor White
Write-Host "---------------------------------------------------------" -ForegroundColor Gray
Write-Host "Para encerrar todos os port-forwards a qualquer momento, execute:" -ForegroundColor Yellow
Write-Host "Stop-Process -Name kubectl -Force" -ForegroundColor White
