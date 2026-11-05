resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "security-group" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "name"
  description = "description"
}

resource "numspot_security_group_rule" "security-group-rule-simple" {
  security_group_id = numspot_security_group.security-group.id
  flow              = "Outbound"
  from_port_range   = 80
  to_port_range     = 80
  ip_protocol       = "tcp"
  ip_range          = "10.0.0.0/16"
}
