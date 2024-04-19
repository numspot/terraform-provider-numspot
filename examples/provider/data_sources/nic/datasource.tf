resource "numspot_vpc" "vpc" {
ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
vpc_id   = numspot_vpc.vpc.id
ip_range = "10.101.1.0/24"
}


resource "numspot_nic" "test" {
subnet_id = numspot_subnet.subnet.id
}

data "numspot_nics" "testdata" {
ids = [numspot_nic.test.id]
depends_on = [numspot_nic.test]
}