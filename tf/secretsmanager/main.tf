module "secretsmanager" {
  source  = "littlejo/secretsmanager/aws"
  version = "VERSION"

  name        = random_pet.name.id
  description = "test"
}

resource "random_pet" "name" {
}
