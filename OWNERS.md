# Owners

> This file is **auto-generated** from [`docs/internal/resources-registry.yaml`](docs/internal/resources-registry.yaml).
> Do not edit by hand: run `make generate-resource-registry`.

## Provider tooling

The **Enabler** team owns the Terraform provider tooling and infrastructure:

- Descriptor patterns and the code scaffold/generation pipeline (`cmd/scaffold`, `cmd/inject-ignores`, `cmd/inject-schema-overrides`, `cmd/resource-registry`).
- The SDK / OpenAPI snapshot pipeline (`internal/sdk`, `make sdk`, `make bump-oas`).
- The provider wiring (`internal/provider`), build tooling (`GNUmakefile`) and CI (`.gitlab-ci.yml`).

Code-owner assignment for these paths is enforced via [`.gitlab/CODEOWNERS`](.gitlab/CODEOWNERS).

## Resources by team

Each Terraform resource is owned by the team declared in its `owner_team` field in the registry.

### cloud-foundations-iaas

- `numspot_bucket` — `internal/services/bucket/`
- `numspot_client_gateway` — `internal/services/clientgateway/`
- `numspot_compute_bridge` — `internal/services/computebridge/`
- `numspot_dhcp_options` — `internal/services/dhcpoptions/`
- `numspot_flexible_gpu` — `internal/services/flexiblegpu/`
- `numspot_hybrid_bridge` — `internal/services/hybridbridge/`
- `numspot_image` — `internal/services/image/`
- `numspot_internet_gateway` — `internal/services/internetgateway/`
- `numspot_keypair` — `internal/services/keypair/`
- `numspot_load_balancer` — `internal/services/loadbalancer/`
- `numspot_nat_gateway` — `internal/services/natgateway/`
- `numspot_nic` — `internal/services/nic/`
- `numspot_public_ip` — `internal/services/publicip/`
- `numspot_route_table` — `internal/services/routetable/`
- `numspot_security_group` — `internal/services/securitygroup/`
- `numspot_security_group_rule` — `internal/services/securitygrouprule/`
- `numspot_server_certificate` — `internal/services/servercertificate/`
- `numspot_snapshot` — `internal/services/snapshot/`
- `numspot_subnet` — `internal/services/subnet/`
- `numspot_virtual_gateway` — `internal/services/virtualgateway/`
- `numspot_vm` — `internal/services/vm/`
- `numspot_volume` — `internal/services/volume/`
- `numspot_vpc` — `internal/services/vpc/`
- `numspot_vpn_connection` — `internal/services/vpnconnection/`

### cloud-foundations-paas

- `numspot_kubernetes_cluster` — `internal/services/kubernetescluster/`
- `numspot_kubernetes_nodepool` — `internal/services/kubernetesnodepool/`
- `numspot_postgres_cluster` — `internal/services/postgrescluster/`

