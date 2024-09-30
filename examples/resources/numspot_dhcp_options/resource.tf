resource "numspot_dhcp_options" "test" {
  domain_name = "foo.bar"
  tags = [
    {
      key   = "Name"
      value = "My DHCP Options"
    }
  ]
}