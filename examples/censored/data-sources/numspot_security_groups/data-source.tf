resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_security_group" "test" {
  net_id      = numspot_vpc.net.id
  name        = "group name"
  description = "this is a security group"
  outbound_rules = [
    {
      from_port_range = 80
      to_port_range   = 80
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    },
    {
      from_port_range = 443
      to_port_range   = 443
      ip_ranges       = ["0.0.0.0/0"]
      ip_protocol     = "tcp"
    }
  ]
}

data "numspot_security_groups" "testdata" {
  ids        = [numspot_security_group.test.id]
  depends_on = [numspot_security_group.test]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_security_groups.testdata.items.0.id"
  }
}