# Create key pair
resource "numspot_key_pair" "create" {
  name = "keypair-example"
}

# Import key pair
resource "numspot_key_pair" "import" {
  name       = "keypair-example"
  public_key = "ssh-rsa ..."
}