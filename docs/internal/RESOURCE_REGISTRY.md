# Référentiel des ressources Terraform Numspot

> Ce document est **généré automatiquement** depuis [`docs/internal/resources-registry.yaml`](resources-registry.yaml).
> Ne pas éditer manuellement : exécuter `make generate-resource-registry`.

## Vue d'ensemble

Ce référentiel centralise, par catégorie de la console Numspot, les informations suivantes pour chaque ressource exposée par le provider Terraform :

- **Async create/delete** : indique si le provider effectue du *state polling* après l'appel API.
- **États possibles (polling)** : états `pending` → `target` surveillés par le provider.
- **Retry pattern** : fonction(s) de retry utilisées.
- **Timeout configurable** : présence d'un champ `timeouts` ou `timeout` dans le schéma Terraform.
- **Backend statuses** : statuts réels du backend Numspot (à compléter depuis les repos backend).

## Légende

- `Oui` : le provider attend que la ressource atteigne un état final.
- `Non` : pas de polling côté provider (opération considérée comme synchrone par le provider).
- `N/A` : opération non applicable ou non implémentée.

## Services Managés

| Ressource | Scope | Owner | Async create | Async delete | États provider (pending → target) | Retry pattern | Timeout configurable | Backend statuses | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `numspot_kubernetes_cluster` | space | numspot-paas | Oui | Non | PENDING, CREATING, DELETING → RUNNING, FAILED | RetryReadUntilStatusStateValidWithTimeout | Oui | _à compléter_ | Seule ressource avec un timeout configurable via `timeouts.create`. |
| `numspot_kubernetes_nodepool` | space | numspot-paas | Non | Non | — | Appel direct + ParseHTTPError | Non | _à compléter_ | Aucun retry ni state polling actuellement. |
| `numspot_postgres_cluster` | space | numspot-paas | Oui | Non | PENDING, CREATING, DELETING → RUNNING, FAILED | RetryReadUntilStatusStateValid | Non | _à compléter_ |  |

## Compute

| Ressource | Scope | Owner | Async create | Async delete | États provider (pending → target) | Retry pattern | Timeout configurable | Backend statuses | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `numspot_flexible_gpu` | space | numspot-iaas | Oui | Non | attaching, detaching → allocated, attached | RetryCreateUntilResourceAvailableWithBody + RetryReadUntilStateValid | Non | _à compléter_ | CRUD logic lives in internal/core/flexible_gpu.go; the resource/datasource delegate to core. |
| `numspot_image` | space | numspot-iaas | Oui | Non | creating, pending → available | RetryCreateUntilResourceAvailableWithBody + RetryReadUntilStateValid | Non | _à compléter_ | scaffold_check enabled; nested bsu deserialization lives in helpers.go (obj.As cannot target the generated attr.Value structs), wired via mapping_overrides.deserialize. |
| `numspot_keypair` | space | numspot-iaas | Non | Non | — | RetryCreateUntilResourceAvailableWithBody | Non | _à compléter_ | scaffold_check enabled; uses id_source mapping (Name is the resource identifier) and ComputedFields for read-only API fields. |
| `numspot_security_group` | space | numspot-iaas | Non | Non | — | RetryCreateUntilResourceAvailableWithBody | Non | _à compléter_ |  |
| `numspot_security_group_rule` | space | numspot-iaas | Non | Non | — | Appel direct + ParseHTTPError | Non | _à compléter_ | Aucun wrapper de retry. |
| `numspot_server_certificate` | space | numspot-iaas | Non | Non | — | Appel direct + ParseHTTPError | Non | _à compléter_ | Aucun wrapper de retry. |
| `numspot_vm` | space | numspot-iaas | Oui | Non | pending, stopping, shutting-down → running, stopped, shutdown, terminated | RetryCreateUntilResourceAvailableWithBody + RetryReadUntilStateValid | Non | _à compléter_ | StopVM et StartVM utilisent aussi RetryReadUntilStateValid. |

## Réseau

| Ressource | Scope | Owner | Async create | Async delete | États provider (pending → target) | Retry pattern | Timeout configurable | Backend statuses | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `numspot_compute_bridge` | space | numspot-iaas | Non | Non | — | Appel direct + ParseHTTPError | Non | _à compléter_ | Pas de wrapper de retry. |
| `numspot_dhcp_options` | space | numspot-iaas | Non | Non | — | RetryCreateUntilResourceAvailableWithBody | Non | _à compléter_ |  |
| `numspot_hybrid_bridge` | space | numspot-iaas | Non | Non | — | Appel direct raw | Non | _à compléter_ | Pas de wrapper de retry. |
| `numspot_internet_gateway` | space | numspot-iaas | Non | Non | — | RetryCreateUntilResourceAvailable | Non | _à compléter_ |  |
| `numspot_load_balancer` | space | numspot-iaas | Non | Non | — | RetryCreateUntilResourceAvailableWithBody | Non | _à compléter_ | Attribut health_check.timeout interne uniquement. |
| `numspot_nat_gateway` | space | numspot-iaas | Oui | Non | pending, deleting → available, deleted | RetryCreateUntilResourceAvailableWithBody + RetryReadUntilStateValid | Non | _à compléter_ |  |
| `numspot_nic` | space | numspot-iaas | Oui | Oui | attaching → attached | RetryCreateUntilResourceAvailableWithBody + StateChangeConf inline (LinkNic.State) | Non | _à compléter_ | Delete ignore silencieusement l'erreur d'unlink. |
| `numspot_public_ip` | space | numspot-iaas | Non | Non | — | RetryCreateUntilResourceAvailable + link direct | Non | _à compléter_ |  |
| `numspot_route_table` | space | numspot-iaas | Non | Non | — | RetryCreateUntilResourceAvailableWithBody | Non | _à compléter_ |  |
| `numspot_subnet` | space | numspot-iaas | Oui | Non | pending → available | RetryCreateUntilResourceAvailableWithBody + RetryReadUntilStateValid | Non | _à compléter_ |  |
| `numspot_vpc` | space | numspot-iaas | Oui | Non | pending, deleting → available | RetryCreateUntilResourceAvailableWithBody + RetryReadUntilStateValid | Non | _à compléter_ |  |

## Stockage

| Ressource | Scope | Owner | Async create | Async delete | États provider (pending → target) | Retry pattern | Timeout configurable | Backend statuses | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `numspot_bucket` | space | numspot-iaas | Non | Non | — | OsClient.CreateBucket (pas de retry wrapper) | Non | _à compléter_ | Object Storage. |
| `numspot_snapshot` | space | numspot-iaas | Oui | Non | pending/queued, in-queue, pending → completed | RetryCreateUntilResourceAvailableWithBody + RetryReadUntilStateValid | Non | _à compléter_ |  |
| `numspot_volume` | space | numspot-iaas | Oui | Non | creating, updating → available, in-use | RetryCreateUntilResourceAvailableWithBody + RetryUntilResourceAvailableWithBody + StateChangeConf inline (LinkedVolumes[0].State) | Non | _à compléter_ |  |

## VPN

| Ressource | Scope | Owner | Async create | Async delete | États provider (pending → target) | Retry pattern | Timeout configurable | Backend statuses | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `numspot_client_gateway` | space | numspot-iaas | Oui | Non | pending → available | RetryCreateUntilResourceAvailableWithBody + RetryReadUntilStateValid | Non | _à compléter_ |  |
| `numspot_virtual_gateway` | space | numspot-iaas | Oui | Non | pending → available | RetryCreateUntilResourceAvailableWithBody + RetryReadUntilStateValid | Non | _à compléter_ |  |
| `numspot_vpn_connection` | space | numspot-iaas | Oui | Non | pending → available | retryCreateVpnCondition custom + RetryReadUntilStateValid | Non | _à compléter_ | Duplique checkRetryCondition + sleep 5s dur. |

---

## Notes complémentaires

- Le timeout par défaut des fonctions de retry est fixé à `5m` pour les opérations API et `15m` pour le polling d'état (`TfRequestStateRetryTimeout`).
- La majorité des suppressions utilisent `RetryDeleteUntilResourceAvailable`, qui est actuellement impacté par le bug B1 du plan de standardisation (codes de retry de création utilisés à la place de ceux de suppression).
- Voir également : [RETRY_ASYNC_STANDARDIZATION_PLAN.md](RETRY_ASYNC_STANDARDIZATION_PLAN.md).
