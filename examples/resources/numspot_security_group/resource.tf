resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "security-group" {
  vpc_id      = numspot_vpc.vpc.id
  name        = "name"
  description = "description"

  tags = [
    {
      key   = "name"
      value = "My Security Group"
    }
  ]
}

# Security group rules are managed with the dedicated
# numspot_security_group_rule resource (see its documentation).
resource "numspot_security_group_rule" "security-group-rule-inbound" {
  security_group_id = numspot_security_group.security-group.id
  flow              = "Inbound"
  from_port_range   = 80
  to_port_range     = 80
  ip_protocol       = "tcp"
  ip_range          = "0.0.0.0/0"
}

resource "numspot_security_group_rule" "security-group-rule-outbound" {
  security_group_id = numspot_security_group.security-group.id
  flow              = "Outbound"
  from_port_range   = 455
  to_port_range     = 455
  ip_protocol       = "tcp"
  ip_range          = "0.0.0.0/0"
}
