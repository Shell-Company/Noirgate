terraform {
  backend "s3" {
    bucket = "noirgate-terraform-state-bucket"
    key    = "shellz.tfstate"
    region = "us-west-2"
  }
}

provider "aws" {
  region = "us-west-2"

}

variable "ImageID" {
  # Ubuntu 18
  default = "ami-0b0e59a09e7f4059f"
}

# SSH RSA key
resource "tls_private_key" "self_signed" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "noirgate-host" {
  key_name   = "noirgate-host"
  public_key = tls_private_key.self_signed.public_key_openssh
}

# Firewall Rules
resource "aws_security_group" "noirgate-host" {
  name = "noirgate-host"
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# instance configuration

resource "aws_instance" "noirgate-host" {
  ami                                  = var.ImageID
  instance_type                        = "t3.large"
  depends_on                           = [aws_key_pair.noirgate-host]
  key_name                             = "noirgate-host"
  security_groups                      = ["${aws_security_group.noirgate-host.name}"]
  instance_initiated_shutdown_behavior = "stop"
  ebs_block_device {
    device_name = "/dev/sda1"
    volume_size = 100
    volume_type = "gp3"
  }


  # Deploy server 
  provisioner "remote-exec" {

    inline = [
      "sudo apt update",
      "sudo apt install docker-compose curl certbot net-tools -y",
      "sudo snap install docker",
      "sudo route add -host 169.254.169.254 reject",
      "sudo docker pull public.ecr.aws/v0z1z7z8/shellcompany/noirgate",
      "sudo docker pull public.ecr.aws/v0z1z7z8/shellcompany/noirgate-manager",
      "sudo docker pull public.ecr.aws/v0z1z7z8/shellcompany/noirgate-discodns",
      "sudo docker pull public.ecr.aws/v0z1z7z8/shellcompany/procurement",
      "git clone https://github.com/shell-company/noirgate-public.git && cd noirgate-public/noirgate-compose && sudo docker-compose up -d",
    ]

    connection {
      type        = "ssh"
      user        = "ubuntu"
      host        = aws_instance.noirgate-host.public_ip
      private_key = tls_private_key.self_signed.private_key_pem
    }
  }

  tags = {
    Name = "noirgate-host"
  }
}


# SSH Stuff
resource "local_file" "provisioned_pem_file" {
  content         = tls_private_key.self_signed.private_key_pem
  filename        = "terraform.pem"
  file_permission = "0400"
}

# Noirgate aws config file
resource "local_file" "noirgate_aws_config" {
  content         = <<EOF
[default]
region = us-west-2
output = json
access_key = ${aws_iam_access_key.noirgate.id}
secret_key = ${aws_iam_access_key.noirgate.secret}
EOF
  filename        = "noirgate_aws_.config"
  file_permission = "0400"
}

output "PublicIP" {
  value = "ssh ubuntu@${aws_instance.noirgate-host.public_ip} -i terraform.pem"
}

output "URL" {
  value = "http://${aws_instance.noirgate-host.public_ip}"
}
