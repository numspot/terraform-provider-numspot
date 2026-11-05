resource "numspot_server_certificate" "crt-terraform-tst-01" {
  name        = "crt-terraform-tst-01"
  body        = file("/path")
  private_key = file("/path")
}
