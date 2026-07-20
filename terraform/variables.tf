variable "aws_region" {
  type        = string
  description = "The AWS region to deploy to"
  default     = "us-east-1"
}

variable "project_name" {
  type        = string
  description = "The name of the project"
  default     = "hw2-staging-v2"
}

variable "vpc_cidr" {
  type        = string
  description = "The CIDR block for the VPC"
  default     = "10.20.0.0/16"
}

variable "public_subnet_cidr" {
  type        = string
  description = "The CIDR block for the public subnet"
  default     = "10.20.1.0/24"
}

variable "instance_type" {
  type        = string
  description = "The EC2 instance type"
  default     = "t3.micro"
}

variable "app_port" {
  type        = number
  description = "The port the application listens on"
  default     = 80
}
