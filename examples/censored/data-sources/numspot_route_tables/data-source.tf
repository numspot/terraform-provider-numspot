resource "numspot_vpc" "net" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.net.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_internet_gateway" "test" {
  vpc_id = numspot_vpc.net.id
}

resource "numspot_route_table" "test" {
  vpc_id    = numspot_vpc.net.id
  subnet_id = numspot_subnet.subnet.id
  routes = [
    {
      destination_ip_range = "0.0.0.0/0"
      gateway_id           = numspot_internet_gateway.test.id
    }
  ]
}

data "numspot_route_tables" "testdata" {
  ids        = [numspot_route_table.test.id]
  depends_on = [numspot_route_table.test]
}

# How to use the datasource in another field
resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_route_tables.testdata.items.0.id"
  }
}