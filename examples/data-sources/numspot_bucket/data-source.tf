resource "numspot_bucket" "bucket" {
  name = "bucket-name"
}

data "numspot_buckets" "datasource-bucket" {
  depends_on = [numspot_bucket.bucket]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_buckets.datasource-bucket.items.0.name"
  }
}