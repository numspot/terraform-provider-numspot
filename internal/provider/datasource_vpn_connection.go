package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpn_connection"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
	"net/http"
)

type vpnConnectionsDataSourceModel struct {
	VPNConnections []resource_vpn_connection.VpnConnectionModel `tfsdk:"vpn_connections"`
	ID             types.String                                 `tfsdk:"id"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &vpnConnectionsDataSource{}
)

func (d *vpnConnectionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*api.ClientWithResponses)
	if !ok || client == nil {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.client = client
}

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewVPNDataSource() datasource.DataSource {
	return &vpnConnectionsDataSource{}
}

// coffeesDataSource is the data source implementation.
type vpnConnectionsDataSource struct {
	client *api.ClientWithResponses
}

// Metadata returns the data source type name.
func (d *vpnConnectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_connections"
}

// Schema defines the schema for the data source.
func (d *vpnConnectionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"vpn_connections": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"client_gateway_configuration": schema.StringAttribute{
							Computed:            true,
							Description:         "Example configuration for the client gateway.",
							MarkdownDescription: "Example configuration for the client gateway.",
						},
						"client_gateway_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the client gateway.",
							MarkdownDescription: "The ID of the client gateway.",
						},
						"connection_type": schema.StringAttribute{
							Computed:            true,
							Description:         "The type of VPN connection (only `ipsec.1` is supported).",
							MarkdownDescription: "The type of VPN connection (only `ipsec.1` is supported).",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the VPN connection.",
							MarkdownDescription: "The ID of the VPN connection.",
						},
						"routes": schema.ListNestedAttribute{
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
								CustomType: resource_vpn_connection.RoutesType{
									ObjectType: types.ObjectType{
										AttrTypes: resource_vpn_connection.RoutesValue{}.AttributeTypes(ctx),
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
										Description:         "The IP on the OUTSCALE side of the tunnel.",
										MarkdownDescription: "The IP on the OUTSCALE side of the tunnel.",
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
								CustomType: resource_vpn_connection.VgwTelemetriesType{
									ObjectType: types.ObjectType{
										AttrTypes: resource_vpn_connection.VgwTelemetriesValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "Information about the current state of one or more of the VPN tunnels.",
							MarkdownDescription: "Information about the current state of one or more of the VPN tunnels.",
						},
						"virtual_gateway_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the virtual gateway.",
							MarkdownDescription: "The ID of the virtual gateway.",
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
									CustomType: resource_vpn_connection.Phase1optionsType{
										ObjectType: types.ObjectType{
											AttrTypes: resource_vpn_connection.Phase1optionsValue{}.AttributeTypes(ctx),
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
											Description:         "The pre-shared key to establish the initial authentication between the client gateway and the virtual gateway. This key can contain any character except line breaks and double quotes (&quot;).",
											MarkdownDescription: "The pre-shared key to establish the initial authentication between the client gateway and the virtual gateway. This key can contain any character except line breaks and double quotes (&quot;).",
										},
									},
									CustomType: resource_vpn_connection.Phase2optionsType{
										ObjectType: types.ObjectType{
											AttrTypes: resource_vpn_connection.Phase2optionsValue{}.AttributeTypes(ctx),
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
							CustomType: resource_vpn_connection.VpnOptionsType{
								ObjectType: types.ObjectType{
									AttrTypes: resource_vpn_connection.VpnOptionsValue{}.AttributeTypes(ctx),
								},
							},
							Computed:            true,
							Description:         "Information about the VPN options.",
							MarkdownDescription: "Information about the VPN options.",
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *vpnConnectionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state vpnConnectionsDataSourceModel
	state.ID = types.StringValue("placeholder")

	res := utils.ExecuteRequest(func() (*api.ReadVpnConnectionsResponse, error) {
		return d.client.ReadVpnConnectionsWithResponse(ctx, spaceID, &api.ReadVpnConnectionsParams{})
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200 == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty load balancers list")
	}

	for _, item := range *res.JSON200.Items {
		tf := VpnConnectionFromHttpToTf(ctx, &item)
		state.VPNConnections = append(state.VPNConnections, tf)
	}
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
