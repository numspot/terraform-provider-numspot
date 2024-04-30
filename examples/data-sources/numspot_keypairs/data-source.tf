resource "numspot_keypair" "test" {
  name = "key-pair-name"
}

data "numspot_keypairs" "testdata" {
  names      = [numspot_keypair.test.name]
  depends_on = [numspot_keypair.test]
}