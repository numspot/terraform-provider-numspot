resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_subnet" "subnet" {
  vpc_id   = numspot_vpc.vpc.id
  ip_range = "10.101.1.0/24"
}

resource "numspot_load_balancer" "lb-test" {
  name = "elb-test"
  listeners = [
    {
      backend_port           = 80
      load_balancer_port     = 80
      load_balancer_protocol = "TCP"

    }
  ]
  subnets = [numspot_subnet.subnet.id]
  type    = "internal"
}
data "numspot_load_balancers" "testdata" {
  load_balancer_names = [numspot_load_balancer.lb-test.name]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_load_balancers.testdata.items.0.id"
  }
}