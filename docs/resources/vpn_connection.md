---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_vpn_connection Resource - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_vpn_connection (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `client_gateway_id` (String) The ID of the client gateway.
- `connection_type` (String) The type of VPN connection (only `ipsec.1` is supported).
- `virtual_gateway_id` (String) The ID of the virtual gateway.

### Optional

- `static_routes_only` (Boolean) If false, the VPN connection uses dynamic routing with Border Gateway Protocol (BGP). If true, routing is controlled using static routes. For more information about how to create and delete static routes, see [CreateVpnConnectionRoute](#createvpnconnectionroute) and [DeleteVpnConnectionRoute](#deletevpnconnectionroute).

### Read-Only

- `client_gateway_configuration` (String) Example configuration for the client gateway.
- `id` (String) The ID of the VPN connection.
- `routes` (Attributes List) Information about one or more static routes associated with the VPN connection, if any. (see [below for nested schema](#nestedatt--routes))
- `state` (String) The state of the VPN connection (`pending` \| `available` \| `deleting` \| `deleted`).
- `vgw_telemetries` (Attributes List) Information about the current state of one or more of the VPN tunnels. (see [below for nested schema](#nestedatt--vgw_telemetries))
- `vpn_options` (Attributes) Information about the VPN options. (see [below for nested schema](#nestedatt--vpn_options))

<a id="nestedatt--routes"></a>
### Nested Schema for `routes`

Read-Only:

- `destination_ip_range` (String) The IP range used for the destination match, in CIDR notation (for example, `10.0.0.0/24`).
- `route_type` (String) The type of route (always `static`).
- `state` (String) The current state of the static route (`pending` \| `available` \| `deleting` \| `deleted`).


<a id="nestedatt--vgw_telemetries"></a>
### Nested Schema for `vgw_telemetries`

Read-Only:

- `accepted_route_count` (Number) The number of routes accepted through BGP (Border Gateway Protocol) route exchanges.
- `last_state_change_date` (String) The date and time (UTC) of the latest state update.
- `outside_ip_address` (String) The IP on the OUTSCALE side of the tunnel.
- `state` (String) The state of the IPSEC tunnel (`UP` \| `DOWN`).
- `state_description` (String) A description of the current state of the tunnel.


<a id="nestedatt--vpn_options"></a>
### Nested Schema for `vpn_options`

Read-Only:

- `phase1options` (Attributes) Information about Phase 1 of the Internet Key Exchange (IKE) negotiation. When Phase 1 finishes successfully, peers proceed to Phase 2 negotiations. (see [below for nested schema](#nestedatt--vpn_options--phase1options))
- `phase2options` (Attributes) Information about Phase 2 of the Internet Key Exchange (IKE) negotiation. (see [below for nested schema](#nestedatt--vpn_options--phase2options))
- `tunnel_inside_ip_range` (String) The range of inside IPs for the tunnel. This must be a /30 CIDR block from the 169.254.254.0/24 range.

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


<a id="nestedatt--vpn_options--phase2options"></a>
### Nested Schema for `vpn_options.phase2options`

Read-Only:

- `phase2dh_group_numbers` (List of Number) The Diffie-Hellman (DH) group numbers allowed for the VPN tunnel for phase 2.
- `phase2encryption_algorithms` (List of String) The encryption algorithms allowed for the VPN tunnel for phase 2.
- `phase2integrity_algorithms` (List of String) The integrity algorithms allowed for the VPN tunnel for phase 2.
- `phase2lifetime_seconds` (Number) The lifetime for phase 2 of the Internet Key Exchange (IKE) negociation process, in seconds.
- `pre_shared_key` (String) The pre-shared key to establish the initial authentication between the client gateway and the virtual gateway. This key can contain any character except line breaks and double quotes (&quot;).