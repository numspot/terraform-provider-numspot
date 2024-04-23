resource "numspot_keypair" "test" {
  name = "key-pair-name"
}

data "numspot_keypair" "testdata" {
  names      = [numspot_keypair.test.name]
  depends_on = [numspot_keypair.test]
}