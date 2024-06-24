data "numspot_roles" "testdata" {
  space_id = "68134f98-205b-4de4-8b85-f6a786ef6481"
  name     = "kubernetes Viewer"
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_roles.testdata.items.0.id"
  }
}
