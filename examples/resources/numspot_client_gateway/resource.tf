resource "numspot_client_gateway" "client-gateway" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.0.1"
  bgp_asn         = 65000
}