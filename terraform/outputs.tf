output "vpc_id" {
  description = "ID da VPC criada"
  value       = aws_vpc.main.id
}

output "db_endpoint" {
  description = "Endpoint de conexão com o banco de dados RDS PostgreSQL"
  value       = aws_db_instance.postgres.endpoint
}

output "eks_cluster_endpoint" {
  description = "Endpoint de comunicação do cluster EKS"
  value       = aws_eks_cluster.main.endpoint
}

output "eks_cluster_name" {
  description = "Nome do cluster EKS"
  value       = aws_eks_cluster.main.name
}

output "kubeconfig_update_command" {
  description = "Comando para atualizar seu kubeconfig local e conectar ao cluster EKS"
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${aws_eks_cluster.main.name}"
}
