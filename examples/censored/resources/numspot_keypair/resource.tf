# Create key pair
resource "numspot_kepair" "create" {
  name = "keypair-example"
}

# Import key pair
resource "numspot_keypair" "import" {
  name       = "keypair-example"
  public_key = "ssh-ed25519 ..."
}