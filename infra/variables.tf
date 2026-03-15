variable "aws_region" {
  default = "ap-south-1"
}

variable "db_username" {
  description = "Database administrator username"
  type        = string
  default     = "trader"
}

variable "db_password" {
  description = "Database administrator password"
  type        = string
  sensitive   = true
}