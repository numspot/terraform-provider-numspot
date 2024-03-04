---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "numspot_load_balancers Data Source - terraform-provider-numspot"
subcategory: ""
description: |-
  
---

# numspot_load_balancers (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `load_balancer_names` (List of String) One or more load balancer names to filter with.

### Read-Only

- `load_balancers` (Attributes List) (see [below for nested schema](#nestedatt--load_balancers))

<a id="nestedatt--load_balancers"></a>
### Nested Schema for `load_balancers`

Read-Only:

- `application_sticky_cookie_policies` (Attributes List) The stickiness policies defined for the load balancer. (see [below for nested schema](#nestedatt--load_balancers--application_sticky_cookie_policies))
- `backend_ips` (List of String) One or more public IPs of back-end VMs.
- `backend_vm_ids` (List of String) One or more IDs of back-end VMs for the load balancer.
- `dns_name` (String) The DNS name of the load balancer.
- `health_check` (Attributes) Information about the health check configuration. (see [below for nested schema](#nestedatt--load_balancers--health_check))
- `id` (String) ID for /loadBalancers
- `listeners` (Attributes List) One or more listeners to create. (see [below for nested schema](#nestedatt--load_balancers--listeners))
- `name` (String) The unique name of the load balancer (32 alphanumeric or hyphen characters maximum, but cannot start or end with a hyphen).
- `net_id` (String) The ID of the Net for the load balancer.
- `public_ip` (String) (internet-facing only) The public IP you want to associate with the load balancer. If not specified, a public IP owned by 3DS OUTSCALE is associated.
- `secured_cookies` (Boolean) Whether secure cookies are enabled for the load balancer.
- `security_groups` (List of String) (Net only) One or more IDs of security groups you want to assign to the load balancer. If not specified, the default security group of the Net is assigned to the load balancer.
- `source_security_group` (Attributes) Information about the source security group of the load balancer, which you can use as part of your inbound rules for your registered VMs.<br />
To only allow traffic from load balancers, add a security group rule that specifies this source security group as the inbound source. (see [below for nested schema](#nestedatt--load_balancers--source_security_group))
- `sticky_cookie_policies` (Attributes List) The policies defined for the load balancer. (see [below for nested schema](#nestedatt--load_balancers--sticky_cookie_policies))
- `subnets` (List of String) (Net only) The ID of the Subnet in which you want to create the load balancer. Regardless of this Subnet, the load balancer can distribute traffic to all Subnets. This parameter is required in a Net.
- `subregion_names` (List of String) The ID of the Subregion in which the load balancer was created.
- `type` (String) The type of load balancer: `internet-facing` or `internal`. Use this parameter only for load balancers in a Net.

<a id="nestedatt--load_balancers--application_sticky_cookie_policies"></a>
### Nested Schema for `load_balancers.application_sticky_cookie_policies`

Read-Only:

- `cookie_name` (String) The name of the application cookie used for stickiness.
- `policy_name` (String) The mnemonic name for the policy being created. The name must be unique within a set of policies for this load balancer.


<a id="nestedatt--load_balancers--health_check"></a>
### Nested Schema for `load_balancers.health_check`

Read-Only:

- `check_interval` (Number) The number of seconds between two pings (between `5` and `600` both included).
- `healthy_threshold` (Number) The number of consecutive successful pings before considering the VM as healthy (between `2` and `10` both included).
- `path` (String) If you use the HTTP or HTTPS protocols, the ping path.
- `port` (Number) The port number (between `1` and `65535`, both included).
- `protocol` (String) The protocol for the URL of the VM (`HTTP` \| `HTTPS` \| `TCP` \| `SSL`).
- `timeout` (Number) The maximum waiting time for a response before considering the VM as unhealthy, in seconds (between `2` and `60` both included).
- `unhealthy_threshold` (Number) The number of consecutive failed pings before considering the VM as unhealthy (between `2` and `10` both included).


<a id="nestedatt--load_balancers--listeners"></a>
### Nested Schema for `load_balancers.listeners`

Read-Only:

- `backend_port` (Number) The port on which the back-end VM is listening (between `1` and `65535`, both included).
- `backend_protocol` (String) The protocol for routing traffic to back-end VMs (`HTTP` \| `HTTPS` \| `TCP` \| `SSL`).
- `load_balancer_port` (Number) The port on which the load balancer is listening (between `1` and `65535`, both included).
- `load_balancer_protocol` (String) The routing protocol (`HTTP` \| `HTTPS` \| `TCP` \| `SSL`).
- `policy_names` (List of String) The names of the policies. If there are no policies enabled, the list is empty.
- `server_certificate_id` (String) The OUTSCALE Resource Name (ORN) of the server certificate. For more information, see [Resource Identifiers > OUTSCALE Resource Names (ORNs)](https://docs.outscale.com/en/userguide/Resource-Identifiers.html#_outscale_resource_names_orns).


<a id="nestedatt--load_balancers--source_security_group"></a>
### Nested Schema for `load_balancers.source_security_group`

Read-Only:

- `security_group_account_id` (String) The account ID of the owner of the security group.
- `security_group_name` (String) The name of the security group.


<a id="nestedatt--load_balancers--sticky_cookie_policies"></a>
### Nested Schema for `load_balancers.sticky_cookie_policies`

Read-Only:

- `cookie_expiration_period` (Number) The time period, in seconds, after which the cookie should be considered stale.<br />
If `1`, the stickiness session lasts for the duration of the browser session.
- `policy_name` (String) The name of the stickiness policy.