variable "aws_region" {
  description = "Região da AWS para provisionamento"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Nome do projeto para identificação e tags"
  type        = string
  default     = "bbb-votacao"
}

variable "environment" {
  description = "Ambiente da infraestrutura"
  type        = string
  default     = "production"
}

variable "db_username" {
  description = "Username do administrador do PostgreSQL"
  type        = string
  default     = "postgres"
}

variable "db_password" {
  description = "Senha do administrador do PostgreSQL"
  type        = string
  sensitive   = true
  default     = "SenhaSuperSegura123!"
}

variable "db_name" {
  description = "Nome do banco de dados inicial"
  type        = string
  default     = "votacao_production"
}

variable "eks_node_instance_types" {
  description = "Tipos de instâncias para os worker nodes do EKS"
  type        = list(string)
  default     = ["t3.medium"]
}
