#!/bin/bash
# run-k6.sh
# Script que executa o teste de carga K6 no cluster Kubernetes.
# Gera o ConfigMap automaticamente a partir do arquivo load_test_7k.js local,
# garantindo que o cluster sempre use a versão mais recente do script.

SCRIPT="${1:-load_test_7k.js}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT_PATH="$SCRIPT_DIR/$SCRIPT"

if [ ! -f "$SCRIPT_PATH" ]; then
  echo "Erro: Script não encontrado: $SCRIPT_PATH"
  exit 1
fi

echo "--------------------------------------------"
echo " Executando teste de carga K6 no Kubernetes"
echo " Script: $SCRIPT"
echo "--------------------------------------------"

# Remove o job anterior se existir
if kubectl get job k6-heavy-test --ignore-not-found 2>/dev/null | grep -q k6-heavy-test; then
  echo ""
  echo "[1/4] Removendo job anterior..."
  kubectl delete job k6-heavy-test --wait=true
else
  echo ""
  echo "[1/4] Nenhum job anterior encontrado."
fi

# Gera o ConfigMap diretamente do arquivo .js local (sem duplicação)
echo ""
echo "[2/4] Gerando ConfigMap a partir de $SCRIPT..."
kubectl create configmap k6-test-script \
  --from-file="load_test_7k.js=$SCRIPT_PATH" \
  --dry-run=client -o yaml | kubectl apply -f -

# Cria o Job
echo ""
echo "[3/4] Criando o Job de carga..."
kubectl apply -f "$SCRIPT_DIR/k8s-job.yaml"

# Aguarda o pod iniciar e exibe os logs em tempo real
echo ""
echo "[4/4] Aguardando o pod iniciar e exibindo logs em tempo real..."
echo "(Pressione Ctrl+C para sair dos logs sem cancelar o teste)"
echo ""

sleep 3
kubectl logs -f job/k6-heavy-test -c k6
