resource "numspot_dhcp_options" "options" {
  domain_name = "domain"
}

resource "numspot_vpc" "example" {
  ip_range            = "10.0.0.0/16"
  dhcp_options_set_id = numspot_dhcp_options.options.id
  tenancy             = "default"
  tags = [
    {
      key   = "name"
      value = "My VPC"
    }
  ]
}