resource "numspot_managed_service_bridge" "managed-service-bridge" {
  source_managed_service_id      = "" // Managed service ID
  destination_managed_service_id = "" // Managed service ID
}

data "numspot_managed_service_bridges" "datasource-managed-service-bridge" {
  depends_on = [numspot_managed_service_bridge.managed-service-bridge.id]
}

resource "null_resource" "print-datasource-id" {
  provisioner "local-exec" {
    command = "echo data.numspot_managed_service_bridges.datasource-managed-service-bridge.items.0.id"
  }
}