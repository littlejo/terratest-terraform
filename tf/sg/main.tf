module "test" {
  source      = "littlejo/security-group/aws"
  version     = "0.2.0"
  name        = "test"
  description = "test"
  ingress = [
    {
      description = "ssh"
      port        = 22
      cidr_blocks = ["0.0.0.0/0", "10.0.0.0/8"]
      protocol    = "tcp"
    },
    {
      description = "dns"
      port        = 53
      cidr_blocks = ["0.0.0.0/0", "10.0.0.0/8"]
      protocol    = "tcp,udp"
    },
  ]
}
