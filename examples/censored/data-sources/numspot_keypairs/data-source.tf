resource "numspot_keypair" "test" {
  name = "key-pair-name"
}

data "numspot_keypairs" "testdata" {
  keypair_names = [numspot_keypair.test.name]
  depends_on    = [numspot_keypair.test]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_keypairs.testdata.items.0.id"
  }
}