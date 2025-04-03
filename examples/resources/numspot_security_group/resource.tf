resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "security-group" {
  vpc_id      = numspot_vpc.vpc.id
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
  outbound_rules = [
    {
      from_port_range = 455
      to_port_range   = 455
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]

  tags = [
    {
      key   = "name"
      value = "My Security Group"
    }
  ]
}