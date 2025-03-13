resource "numspot_client_gateway" "cgw" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "vgw" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.vpc.id
}

resource "numspot_vpn_connection" "vpc_connection" {
  client_gateway_id  = numspot_client_gateway.cgw.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.vgw.id
  static_routes_only = true

  tags = [
    {
      key   = "Name"
      value = "My VPN Connection"
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