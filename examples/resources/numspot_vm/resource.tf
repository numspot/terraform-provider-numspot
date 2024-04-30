resource "numspot_vm" "example" {
  image_id = "ami-12345678"
}

resource "numspot_vm" "example" {
  image_id           = "ami-12345678"
  vm_type            = "tinav5.c1r1p2"
  keypair_name       = "keypair-example"
  security_group_ids = ["sg-12345678"]
  user_data          = "..."
}

# Tags
resource "numspot_vm" "example" {
  image_id = "ami-12345678"

  tags = [
    {
      key   = "Name"
      value = "Terraform-VM"
    }
  ]
}