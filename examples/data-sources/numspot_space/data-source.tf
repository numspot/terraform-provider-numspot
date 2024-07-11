resource "numspot_space" "test" {
  organisation_id = "88888888-4444-4444-4444-cccccccccccc"
  name            = "the space"
  description     = "the description"
}

data "numspot_space" "testdata" {
  space_id        = numspot_space.test.id
  organisation_id = numspot_space.test.organisation_id
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_space.testdata.id"
  }
}