# run-k6.ps1
# Script que executa o teste de carga K6 no cluster Kubernetes.
# Gera o ConfigMap automaticamente a partir do arquivo load_test_7k.js local,
# garantindo que o cluster sempre use a versão mais recente do script.

param(
    [string]$Script = "load_test_7k.js"
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ScriptPath = Join-Path $ScriptDir $Script

if (-not (Test-Path $ScriptPath)) {
    Write-Error "Script nao encontrado: $ScriptPath"
    exit 1
}

Write-Host "--------------------------------------------" -ForegroundColor Cyan
Write-Host " Executando teste de carga K6 no Kubernetes" -ForegroundColor Cyan
Write-Host " Script: $Script" -ForegroundColor Cyan
Write-Host "--------------------------------------------"

# Remove o job anterior se existir
$jobExists = kubectl get job k6-heavy-test --ignore-not-found 2>$null
if ($jobExists) {
    Write-Host "`n[1/4] Removendo job anterior..." -ForegroundColor Yellow
    kubectl delete job k6-heavy-test --wait=true
} else {
    Write-Host "`n[1/4] Nenhum job anterior encontrado." -ForegroundColor Gray
}

# Gera o ConfigMap diretamente do arquivo .js local (sem duplicação)
Write-Host "`n[2/4] Gerando ConfigMap a partir de $Script..." -ForegroundColor Yellow
kubectl create configmap k6-test-script `
    --from-file="load_test_7k.js=$ScriptPath" `
    --dry-run=client -o yaml | kubectl apply -f -

# Cria o Job
Write-Host "`n[3/4] Criando o Job de carga..." -ForegroundColor Yellow
kubectl apply -f "$ScriptDir\k8s-job.yaml"

# Aguarda o pod iniciar e exibe os logs em tempo real
Write-Host "`n[4/4] Aguardando o pod iniciar e exibindo logs em tempo real..." -ForegroundColor Yellow
Write-Host "(Pressione Ctrl+C para sair dos logs sem cancelar o teste)`n" -ForegroundColor Gray

Start-Sleep -Seconds 3
kubectl logs -f job/k6-heavy-test -c k6
