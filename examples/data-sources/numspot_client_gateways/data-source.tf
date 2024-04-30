resource "numspot_client_gateway" "test" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

data "numspot_client_gateways" "testdata" {
  ids        = [numspot_client_gateway.test.id]
  depends_on = [numspot_client_gateway.test]

}