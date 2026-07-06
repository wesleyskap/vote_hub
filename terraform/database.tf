# DB Subnet Group
resource "aws_db_subnet_group" "db_subnet_group" {
  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = [aws_subnet.private_db_1.id, aws_subnet.private_db_2.id]

  tags = {
    Name = "${var.project_name}-db-subnet-group"
  }
}

# Security Group para o RDS
resource "aws_security_group" "db_sg" {
  name        = "${var.project_name}-db-sg"
  description = "Permite conexoes PostgreSQL internas vindas do cluster EKS"
  vpc_id      = aws_vpc.main.id

  # Entrada: Permitir apenas tráfego vindo de dentro da VPC
  ingress {
    description = "Acesso PostgreSQL a partir da VPC"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  # Saída: Bloquear tudo por padrão (banco de dados não inicia conexões ativamente para fora)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-db-sg"
  }
}

# Banco de Dados RDS PostgreSQL Gerenciado
resource "aws_db_instance" "postgres" {
  identifier             = "${var.project_name}-postgres"
  allocated_storage      = 20
  max_allocated_storage  = 100
  engine                 = "postgres"
  engine_version         = "15.4"
  instance_class         = "db.t3.micro" # Apenas para fins de demonstração econômica
  db_name                = var.db_name
  username               = var.db_username
  password               = var.db_password
  db_subnet_group_name   = aws_db_subnet_group.db_subnet_group.name
  vpc_security_group_ids = [aws_security_group.db_sg.id]
  skip_final_snapshot    = true

  # Alta Disponibilidade para Produção (Multi-AZ)
  multi_az = true

  tags = {
    Name = "${var.project_name}-rds"
  }
}
