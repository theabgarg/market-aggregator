# 1. The VPC (The outer boundary)
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = { Name = "aggregator-vpc" }
}

# 2. The Internet Gateway (The door to the outside world)
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.main.id
  tags = { Name = "aggregator-igw" }
}

# 3. Public Subnet (For the Go EC2 instance)
resource "aws_subnet" "public_1a" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "${var.aws_region}a"
  map_public_ip_on_launch = true # Automatically give servers here a public IP!
  tags = { Name = "public-subnet-1a" }
}

# 4. Private Subnets (For the Postgres Database - No public IPs!)
resource "aws_subnet" "private_1a" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "${var.aws_region}a"
  tags = { Name = "private-subnet-1a" }
}

resource "aws_subnet" "private_1b" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.3.0/24"
  availability_zone = "${var.aws_region}b"
  tags = { Name = "private-subnet-1b" }
}

# 5. Route Table (Tell the public subnet how to use the Internet Gateway)
resource "aws_route_table" "public_rt" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0" # Route all external traffic...
    gateway_id = aws_internet_gateway.igw.id # ...to the IGW
  }
}

# Link the route table to the public subnet
resource "aws_route_table_association" "public_assoc" {
  subnet_id      = aws_subnet.public_1a.id
  route_table_id = aws_route_table.public_rt.id
}