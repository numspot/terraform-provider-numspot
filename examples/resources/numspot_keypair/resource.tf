# Create key pair
resource "numspot_kepair" "keypair" {
  name = "keypair-example"
}

# Import key pair
resource "numspot_keypair" "keypair_imported" {
  name       = "keypair-imported"
  public_key = "ssh-ed25519 ..."
}