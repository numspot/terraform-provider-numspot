resource "numspot_server_certificate" "crt-terraform-tst-01" {
  name        = "crt-terraform-tst-01"
  body        = file("/path")
  private_key = file("/path")
}

data "numspot_server_certificate" "datasource-server-certificate" {
  depends_on = [numspot_server_certificate.crt-terraform-tst-01]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_server_certificate.datasource-server-certificate.items.0.name"
  }
}
