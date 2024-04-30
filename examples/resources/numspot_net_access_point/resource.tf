resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/24"
}

resource "numspot_net_access_point" "test" {
  net_id       = numspot_vpc.test.id
  service_name = "com.outscale.cloudgouv-eu-west-1.oos"

  tags = [
    {
      key   = "name"
      value = "net-access-point-name"
    }
  ]
}