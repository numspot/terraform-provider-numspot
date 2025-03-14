// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package vpnconnection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
)

func VpnConnectionDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bgp_asns": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "The Border Gateway Protocol (BGP) Autonomous System Numbers (ASNs) of the connections.",
				MarkdownDescription: "The Border Gateway Protocol (BGP) Autonomous System Numbers (ASNs) of the connections.",
			},
			"client_gateway_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the client gateways.",
				MarkdownDescription: "The IDs of the client gateways.",
			},
			"connection_types": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The types of the VPN connections (only `ipsec.1` is supported).",
				MarkdownDescription: "The types of the VPN connections (only `ipsec.1` is supported).",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the VPN connections.",
				MarkdownDescription: "The IDs of the VPN connections.",
			},
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"client_gateway_configuration": schema.StringAttribute{
							Computed:            true,
							Description:         "Example configuration for the client gateway.",
							MarkdownDescription: "Example configuration for the client gateway.",
						},
						"client_gateway_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the client gateway used on the client end of the connection.",
							MarkdownDescription: "The ID of the client gateway used on the client end of the connection.",
						},
						"connection_type": schema.StringAttribute{
							Computed:            true,
							Description:         "The type of VPN connection (always `ipsec.1`).",
							MarkdownDescription: "The type of VPN connection (always `ipsec.1`).",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the VPN connection.",
							MarkdownDescription: "The ID of the VPN connection.",
						},
						"routes": schema.SetNestedAttribute{ // MANUALLY EDITED : Use Set type instead of List
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"destination_ip_range": schema.StringAttribute{
										Computed:            true,
										Description:         "The IP range used for the destination match, in CIDR notation (for example, `10.0.0.0/24`).",
										MarkdownDescription: "The IP range used for the destination match, in CIDR notation (for example, `10.0.0.0/24`).",
									},
									"route_type": schema.StringAttribute{
										Computed:            true,
										Description:         "The type of route (always `static`).",
										MarkdownDescription: "The type of route (always `static`).",
									},
									"state": schema.StringAttribute{
										Computed:            true,
										Description:         "The current state of the static route (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
										MarkdownDescription: "The current state of the static route (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
									},
								},
								CustomType: RoutesType{
									ObjectType: types.ObjectType{
										AttrTypes: RoutesValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "Information about one or more static routes associated with the VPN connection, if any.",
							MarkdownDescription: "Information about one or more static routes associated with the VPN connection, if any.",
						},
						"state": schema.StringAttribute{
							Computed:            true,
							Description:         "The state of the VPN connection (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
							MarkdownDescription: "The state of the VPN connection (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
						},
						"static_routes_only": schema.BoolAttribute{
							Computed:            true,
							Description:         "If false, the VPN connection uses dynamic routing with Border Gateway Protocol (BGP). If true, routing is controlled using static routes. For more information about how to create and delete static routes, see [CreateVpnConnectionRoute](#createvpnconnectionroute) and [DeleteVpnConnectionRoute](#deletevpnconnectionroute).",
							MarkdownDescription: "If false, the VPN connection uses dynamic routing with Border Gateway Protocol (BGP). If true, routing is controlled using static routes. For more information about how to create and delete static routes, see [CreateVpnConnectionRoute](#createvpnconnectionroute) and [DeleteVpnConnectionRoute](#deletevpnconnectionroute).",
						},
						"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
						"vgw_telemetries": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"accepted_route_count": schema.Int64Attribute{
										Computed:            true,
										Description:         "The number of routes accepted through BGP (Border Gateway Protocol) route exchanges.",
										MarkdownDescription: "The number of routes accepted through BGP (Border Gateway Protocol) route exchanges.",
									},
									"last_state_change_date": schema.StringAttribute{
										Computed:            true,
										Description:         "The date and time (UTC) of the latest state update.",
										MarkdownDescription: "The date and time (UTC) of the latest state update.",
									},
									"outside_ip_address": schema.StringAttribute{
										Computed:            true,
										Description:         "The IP on the NumSpot side of the tunnel.",
										MarkdownDescription: "The IP on the NumSpot side of the tunnel.",
									},
									"state": schema.StringAttribute{
										Computed:            true,
										Description:         "The state of the IPSEC tunnel (`UP` \\| `DOWN`).",
										MarkdownDescription: "The state of the IPSEC tunnel (`UP` \\| `DOWN`).",
									},
									"state_description": schema.StringAttribute{
										Computed:            true,
										Description:         "A description of the current state of the tunnel.",
										MarkdownDescription: "A description of the current state of the tunnel.",
									},
								},
								CustomType: VgwTelemetriesType{
									ObjectType: types.ObjectType{
										AttrTypes: VgwTelemetriesValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "Information about the current state of one or more of the VPN tunnels.",
							MarkdownDescription: "Information about the current state of one or more of the VPN tunnels.",
						},
						"virtual_gateway_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the virtual gateway used on the NumSpot end of the connection.",
							MarkdownDescription: "The ID of the virtual gateway used on the NumSpot end of the connection.",
						},
						"vpn_options": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"phase1options": schema.SingleNestedAttribute{
									Attributes: map[string]schema.Attribute{
										"dpd_timeout_action": schema.StringAttribute{
											Computed:            true,
											Description:         "The action to carry out after a Dead Peer Detection (DPD) timeout occurs.",
											MarkdownDescription: "The action to carry out after a Dead Peer Detection (DPD) timeout occurs.",
										},
										"dpd_timeout_seconds": schema.Int64Attribute{
											Computed:            true,
											Description:         "The maximum waiting time for a Dead Peer Detection (DPD) response before considering the peer as dead, in seconds.",
											MarkdownDescription: "The maximum waiting time for a Dead Peer Detection (DPD) response before considering the peer as dead, in seconds.",
										},
										"ike_versions": schema.ListAttribute{
											ElementType:         types.StringType,
											Computed:            true,
											Description:         "The Internet Key Exchange (IKE) versions allowed for the VPN tunnel.",
											MarkdownDescription: "The Internet Key Exchange (IKE) versions allowed for the VPN tunnel.",
										},
										"phase1dh_group_numbers": schema.ListAttribute{
											ElementType:         types.Int64Type,
											Computed:            true,
											Description:         "The Diffie-Hellman (DH) group numbers allowed for the VPN tunnel for phase 1.",
											MarkdownDescription: "The Diffie-Hellman (DH) group numbers allowed for the VPN tunnel for phase 1.",
										},
										"phase1encryption_algorithms": schema.ListAttribute{
											ElementType:         types.StringType,
											Computed:            true,
											Description:         "The encryption algorithms allowed for the VPN tunnel for phase 1.",
											MarkdownDescription: "The encryption algorithms allowed for the VPN tunnel for phase 1.",
										},
										"phase1integrity_algorithms": schema.ListAttribute{
											ElementType:         types.StringType,
											Computed:            true,
											Description:         "The integrity algorithms allowed for the VPN tunnel for phase 1.",
											MarkdownDescription: "The integrity algorithms allowed for the VPN tunnel for phase 1.",
										},
										"phase1lifetime_seconds": schema.Int64Attribute{
											Computed:            true,
											Description:         "The lifetime for phase 1 of the IKE negotiation process, in seconds.",
											MarkdownDescription: "The lifetime for phase 1 of the IKE negotiation process, in seconds.",
										},
										"replay_window_size": schema.Int64Attribute{
											Computed:            true,
											Description:         "The number of packets in an IKE replay window.",
											MarkdownDescription: "The number of packets in an IKE replay window.",
										},
										"startup_action": schema.StringAttribute{
											Computed:            true,
											Description:         "The action to carry out when establishing tunnels for a VPN connection.",
											MarkdownDescription: "The action to carry out when establishing tunnels for a VPN connection.",
										},
									},
									CustomType: Phase1optionsType{
										ObjectType: types.ObjectType{
											AttrTypes: Phase1optionsValue{}.AttributeTypes(ctx),
										},
									},
									Computed:            true,
									Description:         "Information about Phase 1 of the Internet Key Exchange (IKE) negotiation. When Phase 1 finishes successfully, peers proceed to Phase 2 negotiations. ",
									MarkdownDescription: "Information about Phase 1 of the Internet Key Exchange (IKE) negotiation. When Phase 1 finishes successfully, peers proceed to Phase 2 negotiations. ",
								},
								"phase2options": schema.SingleNestedAttribute{
									Attributes: map[string]schema.Attribute{
										"phase2dh_group_numbers": schema.ListAttribute{
											ElementType:         types.Int64Type,
											Computed:            true,
											Description:         "The Diffie-Hellman (DH) group numbers allowed for the VPN tunnel for phase 2.",
											MarkdownDescription: "The Diffie-Hellman (DH) group numbers allowed for the VPN tunnel for phase 2.",
										},
										"phase2encryption_algorithms": schema.ListAttribute{
											ElementType:         types.StringType,
											Computed:            true,
											Description:         "The encryption algorithms allowed for the VPN tunnel for phase 2.",
											MarkdownDescription: "The encryption algorithms allowed for the VPN tunnel for phase 2.",
										},
										"phase2integrity_algorithms": schema.ListAttribute{
											ElementType:         types.StringType,
											Computed:            true,
											Description:         "The integrity algorithms allowed for the VPN tunnel for phase 2.",
											MarkdownDescription: "The integrity algorithms allowed for the VPN tunnel for phase 2.",
										},
										"phase2lifetime_seconds": schema.Int64Attribute{
											Computed:            true,
											Description:         "The lifetime for phase 2 of the Internet Key Exchange (IKE) negociation process, in seconds.",
											MarkdownDescription: "The lifetime for phase 2 of the Internet Key Exchange (IKE) negociation process, in seconds.",
										},
										"pre_shared_key": schema.StringAttribute{
											Computed:            true,
											Description:         "The pre-shared key to establish the initial authentication between the client gateway and the virtual gateway. This key can contain any character except line breaks and double quotes (\").", // MANUALLY EDITED : parse HTML encoded characters
											MarkdownDescription: "The pre-shared key to establish the initial authentication between the client gateway and the virtual gateway. This key can contain any character except line breaks and double quotes (\").", // MANUALLY EDITED : parse HTML encoded characters
										},
									},
									CustomType: Phase2optionsType{
										ObjectType: types.ObjectType{
											AttrTypes: Phase2optionsValue{}.AttributeTypes(ctx),
										},
									},
									Computed:            true,
									Description:         "Information about Phase 2 of the Internet Key Exchange (IKE) negotiation. ",
									MarkdownDescription: "Information about Phase 2 of the Internet Key Exchange (IKE) negotiation. ",
								},
								"tunnel_inside_ip_range": schema.StringAttribute{
									Computed:            true,
									Description:         "The range of inside IPs for the tunnel. This must be a /30 CIDR block from the 169.254.254.0/24 range.",
									MarkdownDescription: "The range of inside IPs for the tunnel. This must be a /30 CIDR block from the 169.254.254.0/24 range.",
								},
							},
							CustomType: VpnOptionsType{
								ObjectType: types.ObjectType{
									AttrTypes: VpnOptionsValue{}.AttributeTypes(ctx),
								},
							},
							Computed:            true,
							Description:         "Information about the VPN options.",
							MarkdownDescription: "Information about the VPN options.",
						},
					},
				}, // MANUALLY EDITED : Removed CustomType block
				Computed:            true,
				Description:         "Information about one or more VPN connections.",
				MarkdownDescription: "Information about one or more VPN connections.",
			},
			"route_destination_ip_ranges": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The destination IP ranges.",
				MarkdownDescription: "The destination IP ranges.",
			},
			"states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The states of the VPN connections (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
				MarkdownDescription: "The states of the VPN connections (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
			},
			"static_routes_only": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "If false, the VPN connection uses dynamic routing with Border Gateway Protocol (BGP). If true, routing is controlled using static routes. For more information about how to create and delete static routes, see [CreateVpnConnectionRoute](#createvpnconnectionroute) and [DeleteVpnConnectionRoute](#deletevpnconnectionroute).",
				MarkdownDescription: "If false, the VPN connection uses dynamic routing with Border Gateway Protocol (BGP). If true, routing is controlled using static routes. For more information about how to create and delete static routes, see [CreateVpnConnectionRoute](#createvpnconnectionroute) and [DeleteVpnConnectionRoute](#deletevpnconnectionroute).",
			},
			"tag_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The keys of the tags associated with the VPN connections.",
				MarkdownDescription: "The keys of the tags associated with the VPN connections.",
			},
			"tag_values": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The values of the tags associated with the VPN connections.",
				MarkdownDescription: "The values of the tags associated with the VPN connections.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The key/value combination of the tags associated with the VPN connections, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
				MarkdownDescription: "The key/value combination of the tags associated with the VPN connections, in the following format: \"Filters\":{\"Tags\":[\"TAGKEY=TAGVALUE\"]}.", // MANUALLY EDITED : replaced HTML encoded character
			},
			"virtual_gateway_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the virtual gateways.",
				MarkdownDescription: "The IDs of the virtual gateways.",
			},
			// MANUALLY EDITED : remove spaceId
		},
	}
}

// MANUALLY EDITED : Model declaration removed

// MANUALLY EDITED : Functions associated with ItemsType / ItemsValue and Tags removed
