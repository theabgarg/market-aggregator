output "engine_public_ip" {
  description = "The public IP address of your Go aggregator server"
  value       = aws_instance.aggregator_engine.public_ip
}

output "database_endpoint" {
  description = "The internal URL to connect your Go app to Postgres"
  value       = aws_db_instance.postgres.endpoint
}