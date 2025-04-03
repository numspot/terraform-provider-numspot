resource "numspot_keypair" "keypair" {
  name = "key-pair-name"
}

data "numspot_keypairs" "datasource-kaypair" {
  keypair_names = [numspot_keypair.keypair.name]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_keypairs.datasource-kaypair.items.0.id"
  }
}