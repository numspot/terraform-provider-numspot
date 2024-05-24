resource "numspot_dhcp_options" "test" {
  domain_name = "foo.bar"
  tags = [
    {
      key   = "Name"
      value = "Terraform Provider DHCP Options"
    }
  ]
}