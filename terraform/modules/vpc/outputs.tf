output "vpc_id" {
  value       = aws_vpc.this.id
  description = "VPC ID"
}

output "private_subnet_ids" {
  value       = aws_subnet.private[*].id
  description = "IDs of private subnets"
}

output "public_subnet_ids" {
  value       = aws_subnet.public[*].id
  description = "IDs of public subnets"
}

output "availability_zones" {
  value       = var.availability_zones
  description = "Availability zones used"
}
