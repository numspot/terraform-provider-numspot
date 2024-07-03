resource "numspot_dhcp_options" "test" {
  domain_name = "the_domain_name"
}

data "numspot_dhcp_options" "testdata" {
  domain_names = [numspot_dhcp_options.test.domain_name]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_dhcp_options.testdata.items.0.id"
  }
}