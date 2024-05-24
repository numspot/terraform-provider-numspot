resource "numspot_flexible_gpu" "test" {
  model_name             = "nvidia-a100-80"
  generation             = "v6"
  availability_zone_name = "eu-west-2a"
}