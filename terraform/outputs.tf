output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.main.id
}

output "public_subnet_id" {
  description = "The ID of the public subnet"
  value       = aws_subnet.public.id
}

output "instance_id" {
  description = "The ID of the EC2 instance"
  value       = aws_instance.app.id
}

output "instance_public_ip" {
  description = "The public IP address of the EC2 instance"
  value       = aws_instance.app.public_ip
}

output "application_url" {
  description = "The URL to access the deployed application"
  value       = "http://${aws_instance.app.public_ip}"
}
