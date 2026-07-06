# Terraform - Infraestrutura como código (IaC)

Este diretório contém a especificação da infraestrutura AWS para rodar o sistema de votação do BBB em produção.

## Recursos provisionados
- **VPC** com subnets públicas (para Load Balancers) e privadas de aplicação e banco de dados distribuídas em múltiplas Zonas de Disponibilidade (Multi-AZ).
- **NAT Gateway** e **Internet Gateway** configurados para permitir saída externa segura das subnets privadas.
- **Amazon RDS PostgreSQL** gerenciado e de alta disponibilidade (Multi-AZ) isolado na rede privada.
- **Amazon EKS Cluster** (Control Plane) e **Node Group** auto-escalável configurados para abrigar nossos microsserviços.

---

## Como utilizar

### Instalação e autenticação
Antes de iniciar, garanta que você tem instalado localmente:
- [Terraform CLI](https://developer.hashicorp.com/terraform/downloads)
- [AWS CLI](https://aws.amazon.com/cli/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

Configure suas credenciais da AWS rodando:
```bash
aws configure
```

### Comandos do Terraform
Dentro da pasta `terraform/`, execute os passos a seguir:

#### Inicializar módulos
Baixa o driver da AWS e os módulos necessários.
```bash
terraform init
```

#### Validação sintática
Valida se não existem erros de sintaxe ou de referência no código.
```bash
terraform validate
```

#### Visualizar plano
Mostra o que será criado, modificado ou destruído na nuvem.
```bash
terraform plan
```

#### Aplicar infraestrutura
Aplica as mudanças reais na sua conta AWS.
```bash
terraform apply -auto-approve
```

---

## Conectando ao cluster EKS e executando o deploy

Depois que o `terraform apply` terminar, configure o acesso ao Kubernetes com o comando abaixo:

```bash
aws eks update-kubeconfig --region us-east-1 --name bbb-votacao-cluster
```

Valide se o cluster está respondendo:
```bash
kubectl get nodes
```

Agora é só aplicar os manifestos da pasta `k8s/`:
```bash
kubectl apply -f ../k8s/
```
