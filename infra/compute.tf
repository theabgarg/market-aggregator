# 1. The Data Source (This finds the latest Amazon Linux OS ID dynamically)
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]
  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }
}

# 2. The Actual Server
resource "aws_instance" "aggregator_engine" {
  # This references the data block above!
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"

  subnet_id              = aws_subnet.public_1a.id
  vpc_security_group_ids = [aws_security_group.engine_sg.id]

  tags = { Name = "Go-Market-Engine" }

  # The Bash script that runs on first boot
  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              yum install -y docker git
              systemctl start docker
              systemctl enable docker
              usermod -aG docker ec2-user
              docker run -d -p 8080:80 nginxdemos/hello
              EOF

  user_data_replace_on_change = true 
}