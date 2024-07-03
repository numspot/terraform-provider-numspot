resource "numspot_vm" "example" {
  image_id = "ami-12345678"
}

resource "numspot_vm" "example" {
  image_id           = "ami-12345678"
  type               = "ns-cus6-2c4r"
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