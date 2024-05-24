data "numspot_product_types" "testdata" {
  ids = ["0001"]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_product_types.testdata.items.0.id"
  }
}