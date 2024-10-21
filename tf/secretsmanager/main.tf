module "secretsmanager" {
  source  = "littlejo/secretsmanager/aws"
  version = "0.2.0"

  name        = "test"
  description = "test"
}
