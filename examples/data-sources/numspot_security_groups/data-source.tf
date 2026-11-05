resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "security-group" {
  net_id      = numspot_vpc.vpc.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 443
      to_port_range   = 443
      ip_protocol     = "tcp"
    }
  ]
}

data "numspot_security_groups" "datasource-security-group" {
  security_group_ids = [numspot_security_group.security-group.id]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_security_groups.datasource-security-group.items.0.id"
  }
}
