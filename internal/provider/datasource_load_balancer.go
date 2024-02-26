package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_load_balancer"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
	"net/http"
)

type loadBalancersDataSourceModel struct {
	LoadBalancers []resource_load_balancer.LoadBalancerModel `tfsdk:"load_balancers"`
	ID            types.String                               `tfsdk:"id"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &loadBalancersDataSource{}
)

func (d *loadBalancersDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
func NewLoadBalancersDataSource() datasource.DataSource {
	return &loadBalancersDataSource{}
}

// coffeesDataSource is the data source implementation.
type loadBalancersDataSource struct {
	client *api.ClientWithResponses
}

// Metadata returns the data source type name.
func (d *loadBalancersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancers"
}

// Schema defines the schema for the data source.
func (d *loadBalancersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"load_balancers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"application_sticky_cookie_policies": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"cookie_name": schema.StringAttribute{
										Computed:            true,
										Description:         "The name of the application cookie used for stickiness.",
										MarkdownDescription: "The name of the application cookie used for stickiness.",
									},
									"policy_name": schema.StringAttribute{
										Computed:            true,
										Description:         "The mnemonic name for the policy being created. The name must be unique within a set of policies for this load balancer.",
										MarkdownDescription: "The mnemonic name for the policy being created. The name must be unique within a set of policies for this load balancer.",
									},
								},
								CustomType: resource_load_balancer.ApplicationStickyCookiePoliciesType{
									ObjectType: types.ObjectType{
										AttrTypes: resource_load_balancer.ApplicationStickyCookiePoliciesValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "The stickiness policies defined for the load balancer.",
							MarkdownDescription: "The stickiness policies defined for the load balancer.",
						},
						"backend_ips": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "One or more public IPs of back-end VMs.",
							MarkdownDescription: "One or more public IPs of back-end VMs.",
						},
						"backend_vm_ids": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "One or more IDs of back-end VMs for the load balancer.",
							MarkdownDescription: "One or more IDs of back-end VMs for the load balancer.",
						},
						"dns_name": schema.StringAttribute{
							Computed:            true,
							Description:         "The DNS name of the load balancer.",
							MarkdownDescription: "The DNS name of the load balancer.",
						},
						"health_check": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"check_interval": schema.Int64Attribute{
									Computed:            true,
									Description:         "The number of seconds between two pings (between `5` and `600` both included).",
									MarkdownDescription: "The number of seconds between two pings (between `5` and `600` both included).",
								},
								"healthy_threshold": schema.Int64Attribute{
									Computed:            true,
									Description:         "The number of consecutive successful pings before considering the VM as healthy (between `2` and `10` both included).",
									MarkdownDescription: "The number of consecutive successful pings before considering the VM as healthy (between `2` and `10` both included).",
								},
								"path": schema.StringAttribute{
									Computed:            true,
									Description:         "If you use the HTTP or HTTPS protocols, the ping path.",
									MarkdownDescription: "If you use the HTTP or HTTPS protocols, the ping path.",
								},
								"port": schema.Int64Attribute{
									Computed:            true,
									Description:         "The port number (between `1` and `65535`, both included).",
									MarkdownDescription: "The port number (between `1` and `65535`, both included).",
								},
								"protocol": schema.StringAttribute{
									Computed:            true,
									Description:         "The protocol for the URL of the VM (`HTTP` \\| `HTTPS` \\| `TCP` \\| `SSL`).",
									MarkdownDescription: "The protocol for the URL of the VM (`HTTP` \\| `HTTPS` \\| `TCP` \\| `SSL`).",
								},
								"timeout": schema.Int64Attribute{
									Computed:            true,
									Description:         "The maximum waiting time for a response before considering the VM as unhealthy, in seconds (between `2` and `60` both included).",
									MarkdownDescription: "The maximum waiting time for a response before considering the VM as unhealthy, in seconds (between `2` and `60` both included).",
								},
								"unhealthy_threshold": schema.Int64Attribute{
									Computed:            true,
									Description:         "The number of consecutive failed pings before considering the VM as unhealthy (between `2` and `10` both included).",
									MarkdownDescription: "The number of consecutive failed pings before considering the VM as unhealthy (between `2` and `10` both included).",
								},
							},
							CustomType: resource_load_balancer.HealthCheckType{
								ObjectType: types.ObjectType{
									AttrTypes: resource_load_balancer.HealthCheckValue{}.AttributeTypes(ctx),
								},
							},
							Computed:            true,
							Description:         "Information about the health check configuration.",
							MarkdownDescription: "Information about the health check configuration.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "ID for /loadBalancers",
							MarkdownDescription: "ID for /loadBalancers",
						},
						"listeners": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"backend_port": schema.Int64Attribute{
										Computed:            true,
										Description:         "The port on which the back-end VM is listening (between `1` and `65535`, both included).",
										MarkdownDescription: "The port on which the back-end VM is listening (between `1` and `65535`, both included).",
									},
									"backend_protocol": schema.StringAttribute{
										Computed:            true,
										Description:         "The protocol for routing traffic to back-end VMs (`HTTP` \\| `HTTPS` \\| `TCP` \\| `SSL`).",
										MarkdownDescription: "The protocol for routing traffic to back-end VMs (`HTTP` \\| `HTTPS` \\| `TCP` \\| `SSL`).",
									},
									"load_balancer_port": schema.Int64Attribute{
										Computed:            true,
										Description:         "The port on which the load balancer is listening (between `1` and `65535`, both included).",
										MarkdownDescription: "The port on which the load balancer is listening (between `1` and `65535`, both included).",
									},
									"load_balancer_protocol": schema.StringAttribute{
										Computed:            true,
										Description:         "The routing protocol (`HTTP` \\| `HTTPS` \\| `TCP` \\| `SSL`).",
										MarkdownDescription: "The routing protocol (`HTTP` \\| `HTTPS` \\| `TCP` \\| `SSL`).",
									},
									"policy_names": schema.ListAttribute{
										ElementType:         types.StringType,
										Computed:            true,
										Description:         "The names of the policies. If there are no policies enabled, the list is empty.",
										MarkdownDescription: "The names of the policies. If there are no policies enabled, the list is empty.",
									},
									"server_certificate_id": schema.StringAttribute{
										Computed:            true,
										Description:         "The OUTSCALE Resource Name (ORN) of the server certificate. For more information, see [Resource Identifiers > OUTSCALE Resource Names (ORNs)](https://docs.outscale.com/en/userguide/Resource-Identifiers.html#_outscale_resource_names_orns).",
										MarkdownDescription: "The OUTSCALE Resource Name (ORN) of the server certificate. For more information, see [Resource Identifiers > OUTSCALE Resource Names (ORNs)](https://docs.outscale.com/en/userguide/Resource-Identifiers.html#_outscale_resource_names_orns).",
									},
								},
								CustomType: resource_load_balancer.ListenersType{
									ObjectType: types.ObjectType{
										AttrTypes: resource_load_balancer.ListenersValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "One or more listeners to create.",
							MarkdownDescription: "One or more listeners to create.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "The unique name of the load balancer (32 alphanumeric or hyphen characters maximum, but cannot start or end with a hyphen).",
							MarkdownDescription: "The unique name of the load balancer (32 alphanumeric or hyphen characters maximum, but cannot start or end with a hyphen).",
						},
						"net_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the Net for the load balancer.",
							MarkdownDescription: "The ID of the Net for the load balancer.",
						},
						"public_ip": schema.StringAttribute{
							Computed:            true,
							Description:         "(internet-facing only) The public IP you want to associate with the load balancer. If not specified, a public IP owned by 3DS OUTSCALE is associated.",
							MarkdownDescription: "(internet-facing only) The public IP you want to associate with the load balancer. If not specified, a public IP owned by 3DS OUTSCALE is associated.",
						},
						"secured_cookies": schema.BoolAttribute{
							Computed:            true,
							Description:         "Whether secure cookies are enabled for the load balancer.",
							MarkdownDescription: "Whether secure cookies are enabled for the load balancer.",
						},
						"security_groups": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "(Net only) One or more IDs of security groups you want to assign to the load balancer. If not specified, the default security group of the Net is assigned to the load balancer.",
							MarkdownDescription: "(Net only) One or more IDs of security groups you want to assign to the load balancer. If not specified, the default security group of the Net is assigned to the load balancer.",
						},
						"source_security_group": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"security_group_account_id": schema.StringAttribute{
									Computed:            true,
									Description:         "The account ID of the owner of the security group.",
									MarkdownDescription: "The account ID of the owner of the security group.",
								},
								"security_group_name": schema.StringAttribute{
									Computed:            true,
									Description:         "The name of the security group.",
									MarkdownDescription: "The name of the security group.",
								},
							},
							CustomType: resource_load_balancer.SourceSecurityGroupType{
								ObjectType: types.ObjectType{
									AttrTypes: resource_load_balancer.SourceSecurityGroupValue{}.AttributeTypes(ctx),
								},
							},
							Computed:            true,
							Description:         "Information about the source security group of the load balancer, which you can use as part of your inbound rules for your registered VMs.<br />\nTo only allow traffic from load balancers, add a security group rule that specifies this source security group as the inbound source.",
							MarkdownDescription: "Information about the source security group of the load balancer, which you can use as part of your inbound rules for your registered VMs.<br />\nTo only allow traffic from load balancers, add a security group rule that specifies this source security group as the inbound source.",
						},
						"sticky_cookie_policies": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"cookie_expiration_period": schema.Int64Attribute{
										Computed:            true,
										Description:         "The time period, in seconds, after which the cookie should be considered stale.<br />\nIf `1`, the stickiness session lasts for the duration of the browser session.",
										MarkdownDescription: "The time period, in seconds, after which the cookie should be considered stale.<br />\nIf `1`, the stickiness session lasts for the duration of the browser session.",
									},
									"policy_name": schema.StringAttribute{
										Computed:            true,
										Description:         "The name of the stickiness policy.",
										MarkdownDescription: "The name of the stickiness policy.",
									},
								},
								CustomType: resource_load_balancer.StickyCookiePoliciesType{
									ObjectType: types.ObjectType{
										AttrTypes: resource_load_balancer.StickyCookiePoliciesValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "The policies defined for the load balancer.",
							MarkdownDescription: "The policies defined for the load balancer.",
						},
						"subnets": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "(Net only) The ID of the Subnet in which you want to create the load balancer. Regardless of this Subnet, the load balancer can distribute traffic to all Subnets. This parameter is required in a Net.",
							MarkdownDescription: "(Net only) The ID of the Subnet in which you want to create the load balancer. Regardless of this Subnet, the load balancer can distribute traffic to all Subnets. This parameter is required in a Net.",
						},
						"subregion_names": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "The ID of the Subregion in which the load balancer was created.",
							MarkdownDescription: "The ID of the Subregion in which the load balancer was created.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							Description:         "The type of load balancer: `internet-facing` or `internal`. Use this parameter only for load balancers in a Net.",
							MarkdownDescription: "The type of load balancer: `internet-facing` or `internal`. Use this parameter only for load balancers in a Net.",
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *loadBalancersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state loadBalancersDataSourceModel
	state.ID = types.StringValue("placeholder")

	res := utils.ExecuteRequest(func() (*api.ReadLoadBalancersResponse, error) {
		return d.client.ReadLoadBalancersWithResponse(ctx, &api.ReadLoadBalancersParams{})
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.LoadBalancers == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty load balancers list")
	}

	for _, item := range *res.JSON200.LoadBalancers {
		tf := LoadBalancerFromHttpToTf(ctx, &item)
		state.LoadBalancers = append(state.LoadBalancers, tf)
	}
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
