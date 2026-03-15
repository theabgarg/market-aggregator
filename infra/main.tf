# infra/main.tf

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

# This is the block OpenTofu is crying about! 
# It tells OpenTofu exactly which region to use.
provider "aws" {
  region = var.aws_region 
  
  default_tags {
    tags = {
      Project     = "MarketAggregator"
      Environment = "Production"
    }
  }
}