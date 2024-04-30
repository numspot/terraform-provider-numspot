resource "numspot_security_group" "test" {
  net_id      = numspot_vpc.net.id
  name        = "name"
  description = "description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

# Tags
resource "numspot_security_group" "test" {
  net_id      = numspot_vpc.net.id
  name        = "name"
  description = "description"

  inbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]

  tags = [
    {
      key   = "name"
      value = "security-group-name"
    }
  ]
}