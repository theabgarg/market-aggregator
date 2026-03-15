# 1. EC2 Security Group
resource "aws_security_group" "engine_sg" {
  name        = "engine-sg"
  description = "Allow SSH and API access"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "Allow SSH from anywhere (For you to log in)"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] 
  }

  ingress {
    description = "Allow WebSocket/API traffic"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Allow the engine to talk to the internet (Binance, Docker Hub)"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# 2. Database Security Group
resource "aws_security_group" "db_sg" {
  name        = "database-sg"
  description = "Allow Postgres access ONLY from the Engine"
  vpc_id      = aws_vpc.main.id

  ingress {
    description     = "Allow traffic from the engine SG"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    # THE MAGIC: Instead of an IP address, we whitelist the Engine's Security Group!
    security_groups = [aws_security_group.engine_sg.id]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}