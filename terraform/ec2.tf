data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }
}

resource "aws_iam_role" "app_role" {
  name = "${var.project_name}-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_instance_profile" "app_profile" {
  name = "${var.project_name}-ec2-profile"
  role = aws_iam_role.app_role.name
}

resource "aws_instance" "app" {
  ami                         = data.aws_ami.amazon_linux.id
  instance_type               = var.instance_type
  subnet_id                   = aws_subnet.public.id
  associate_public_ip_address = true
  ebs_optimized               = true
  monitoring                  = true
  iam_instance_profile        = aws_iam_instance_profile.app_profile.name

  vpc_security_group_ids = [aws_security_group.app_sg.id]

  root_block_device {
    encrypted = true
  }

  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }

  user_data = <<-EOF
#!/bin/bash
set -e

# Update packages and install golang
dnf update -y
dnf install -y golang git

# Create app directory
mkdir -p /opt/app
cd /opt/app

# The app code will be deployed here. We embed it using terraform file function.
cat << 'APP_EOF' > main.go
${file("${path.module}/../app/main.go")}
APP_EOF

go mod init nsp-app
go build -o server main.go

# Setup systemd service
cat << 'SVC_EOF' > /etc/systemd/system/nsp-app.service
[Unit]
Description=NSP-4-S2-S25App Staging Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/app
ExecStart=/opt/app/server
Restart=on-failure

[Install]
WantedBy=multi-user.target
SVC_EOF

systemctl daemon-reload
systemctl enable nsp-app
systemctl start nsp-app
  EOF

  tags = {
    Name = "${var.project_name}-instance"
  }
}
