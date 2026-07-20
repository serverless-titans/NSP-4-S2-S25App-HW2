resource "aws_security_group" "app_sg" {
  name        = "${var.project_name}-app-sg"
  description = "Security group for the staging application"
  vpc_id      = aws_vpc.main.id

  # checkov:skip=CKV_AWS_260: Assignment requires public access on port 80 for demonstration
  ingress {
    description = "Allow HTTP inbound traffic"
    from_port   = var.app_port
    to_port     = var.app_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # checkov:skip=CKV_AWS_382: Allow all egress for downloading packages and updates
  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-app-sg"
  }
}
