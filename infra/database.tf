# Group our private subnets together for the RDS instance
resource "aws_db_subnet_group" "db_subnets" {
  name       = "main-db-subnet-group"
  subnet_ids = [aws_subnet.private_1a.id, aws_subnet.private_1b.id]
}

resource "aws_db_instance" "postgres" {
  identifier             = "market-data-db"
  engine                 = "postgres"
  engine_version         = "17.9"
  instance_class         = "db.t3.micro" # Free Tier Eligible
  allocated_storage      = 20            # Free Tier Eligible
  db_name                = "market_data"
  username               = var.db_username
  password               = var.db_password
  db_subnet_group_name   = aws_db_subnet_group.db_subnets.name
  vpc_security_group_ids = [aws_security_group.db_sg.id]
  skip_final_snapshot    = true          # Essential so Terraform can delete it cleanly without hanging
  publicly_accessible    = false         # Highly secure
}