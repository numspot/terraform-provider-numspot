resource "numspot_public_ip" "test" {}

data "numspot_public_ip" "testdata" {
  ids        = [numspot_public_ip.test.id]
  depends_on = [numspot_public_ip.test]
}