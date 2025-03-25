resource "numspot_flexible_gpu" "test" {
  model_name             = "nvidia-p100"
  generation             = "v5"
  availability_zone_name = "eu-west-2a"
}

data "numspot_flexible_gpus" "testdata" {
  ids = [numspot_flexible_gpu.test.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_flexible_gpus.testdata.items.0.id"
  }
}