resource "numspot_dhcp_options" "dhcp-options" {
  domain_name = "foo.bar"
  tags = [
    {
      key   = "Name"
      value = "My DHCP Options"
    }
  ]
}