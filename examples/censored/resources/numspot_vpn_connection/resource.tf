resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

resource "numspot_vpc" "test" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "test" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.test.id
}

resource "numspot_vpn_connection" "test" {
  client_gateway_id  = numspot_client_gateway.test.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.test.id
  static_routes_only = true

  tags = [
    {
      key   = "Name"
      value = "Terraform VPN Connection"
    }
  ]
  vpn_options = {
    phase2options = {
      pre_shared_key = "sample key !"
    }
    tunnel_inside_ip_range = "169.254.254.22/30"
  }
  routes = [
    {
      destination_ip_range = "192.0.2.0/24"
    },
    {
      destination_ip_range = "192.168.255.0/24"
    }
  ]
}