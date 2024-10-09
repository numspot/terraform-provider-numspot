---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_vpn_connection Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_vpn_connection (Resource)



## Example Usage

```terraform
resource "numspot_client_gateway" "cgw" {
  connection_type = "ipsec.1"
  public_ip       = "192.0.2.0"
  bgp_asn         = 65000
}

resource "numspot_vpc" "vpc" {
  ip_range = "10.101.0.0/16"
}

resource "numspot_virtual_gateway" "vgw" {
  connection_type = "ipsec.1"
  vpc_id          = numspot_vpc.vpc.id
}

resource "numspot_vpn_connection" "vpc_connection" {
  client_gateway_id  = numspot_client_gateway.cgw.id
  connection_type    = "ipsec.1"
  virtual_gateway_id = numspot_virtual_gateway.vgw.id
  static_routes_only = true

  tags = [
    {
      key   = "Name"
      value = "My VPN Connection"
    }
  ]
  vpn_options = {
    phase2options = {
      pre_shared_key = "sample key !"
    }
    tunnel_inside_ip_range = "169.254.254.22/30"
  }
  routes = [
    {
      destination_ip_range = "192.0.2.0/24"
    },
    {
      destination_ip_range = "192.168.255.0/24"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `client_gateway_id` (String) The ID of the client gateway.
- `connection_type` (String) The type of VPN connection (only `ipsec.1` is supported).
- `virtual_gateway_id` (String) The ID of the virtual gateway.

### Optional

- `routes` (Attributes Set) Information about one or more static routes associated with the VPN connection, if any. (see [below for nested schema](#nestedatt--routes))
- `static_routes_only` (Boolean) By default or if false, the VPN connection uses dynamic routing with Border Gateway Protocol (BGP). If true, routing is controlled using static routes. For more information about how to create and delete static routes, see [CreateVpnConnectionRoute](#createvpnconnectionroute) and [DeleteVpnConnectionRoute](#deletevpnconnectionroute).
- `tags` (Attributes List) One or more tags associated with the resource. (see [below for nested schema](#nestedatt--tags))
- `vpn_options` (Attributes) Information about the VPN options. (see [below for nested schema](#nestedatt--vpn_options))

### Read-Only

- `client_gateway_configuration` (String) Example configuration for the client gateway.
- `id` (String) The ID of the VPN connection.
- `state` (String) The state of the VPN connection (`pending` \| `available` \| `deleting` \| `deleted`).
- `vgw_telemetries` (Attributes List) Information about the current state of one or more of the VPN tunnels. (see [below for nested schema](#nestedatt--vgw_telemetries))

<a id="nestedatt--routes"></a>
### Nested Schema for `routes`

Optional:

- `destination_ip_range` (String) The IP range used for the destination match, in CIDR notation (for example, `10.0.0.0/24`).

Read-Only:

- `route_type` (String) The type of route (always `static`).
- `state` (String) The current state of the static route (`pending` \| `available` \| `deleting` \| `deleted`).


<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Required:

- `key` (String) The key of the tag, with a minimum of 1 character.
- `value` (String) The value of the tag, between 0 and 255 characters.


<a id="nestedatt--vpn_options"></a>
### Nested Schema for `vpn_options`

Optional:

- `phase2options` (Attributes) Information about Phase 2 of the Internet Key Exchange (IKE) negotiation. (see [below for nested schema](#nestedatt--vpn_options--phase2options))
- `tunnel_inside_ip_range` (String) The range of inside IPs for the tunnel. This must be a /30 CIDR block from the 169.254.254.0/24 range.

Read-Only:

- `phase1options` (Attributes) Information about Phase 1 of the Internet Key Exchange (IKE) negotiation. When Phase 1 finishes successfully, peers proceed to Phase 2 negotiations. (see [below for nested schema](#nestedatt--vpn_options--phase1options))

<a id="nestedatt--vpn_options--phase2options"></a>
### Nested Schema for `vpn_options.phase2options`

Optional:

- `pre_shared_key` (String) The pre-shared key to establish the initial authentication between the client gateway and the virtual gateway. This key can contain any character except line breaks and double quotes (").

Read-Only:

- `phase2dh_group_numbers` (List of Number) The Diffie-Hellman (DH) group numbers allowed for the VPN tunnel for phase 2.
- `phase2encryption_algorithms` (List of String) The encryption algorithms allowed for the VPN tunnel for phase 2.
- `phase2integrity_algorithms` (List of String) The integrity algorithms allowed for the VPN tunnel for phase 2.
- `phase2lifetime_seconds` (Number) The lifetime for phase 2 of the Internet Key Exchange (IKE) negociation process, in seconds.


<a id="nestedatt--vpn_options--phase1options"></a>
### Nested Schema for `vpn_options.phase1options`

Read-Only:

- `dpd_timeout_action` (String) The action to carry out after a Dead Peer Detection (DPD) timeout occurs.
- `dpd_timeout_seconds` (Number) The maximum waiting time for a Dead Peer Detection (DPD) response before considering the peer as dead, in seconds.
- `ike_versions` (List of String) The Internet Key Exchange (IKE) versions allowed for the VPN tunnel.
- `phase1dh_group_numbers` (List of Number) The Diffie-Hellman (DH) group numbers allowed for the VPN tunnel for phase 1.
- `phase1encryption_algorithms` (List of String) The encryption algorithms allowed for the VPN tunnel for phase 1.
- `phase1integrity_algorithms` (List of String) The integrity algorithms allowed for the VPN tunnel for phase 1.
- `phase1lifetime_seconds` (Number) The lifetime for phase 1 of the IKE negotiation process, in seconds.
- `replay_window_size` (Number) The number of packets in an IKE replay window.
- `startup_action` (String) The action to carry out when establishing tunnels for a VPN connection.



<a id="nestedatt--vgw_telemetries"></a>
### Nested Schema for `vgw_telemetries`

Read-Only:

- `accepted_route_count` (Number) The number of routes accepted through BGP (Border Gateway Protocol) route exchanges.
- `last_state_change_date` (String) The date and time (UTC) of the latest state update.
- `outside_ip_address` (String) The IP on the NumSpot side of the tunnel.
- `state` (String) The state of the IPSEC tunnel (`UP` \| `DOWN`).
- `state_description` (String) A description of the current state of the tunnel.