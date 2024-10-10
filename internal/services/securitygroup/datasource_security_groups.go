package securitygroup

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type SecurityGroupsDataSourceModel struct {
	Items                          []SecurityGroupModel `tfsdk:"items"`
	Descriptions                   types.List           `tfsdk:"descriptions"`
	InboundRuleFromPortRanges      types.List           `tfsdk:"inbound_rule_from_port_ranges"`
	InboundRuleIpRanges            types.List           `tfsdk:"inbound_rule_ip_ranges"`
	InboundRuleProtocols           types.List           `tfsdk:"inbound_rule_protocols"`
	InboundRuleSecurityGroupIds    types.List           `tfsdk:"inbound_rule_security_group_ids"`
	InboundRuleSecurityGroupNames  types.List           `tfsdk:"inbound_rule_security_group_names"`
	InboundRuleToPortRanges        types.List           `tfsdk:"inbound_rule_to_port_ranges"`
	OutboundRuleFromPortRanges     types.List           `tfsdk:"outbound_rule_from_port_ranges"`
	OutboundRuleIpRanges           types.List           `tfsdk:"outbound_rule_ip_ranges"`
	OutboundRuleProtocols          types.List           `tfsdk:"outbound_rule_protocols"`
	OutboundRuleSecurityGroupIds   types.List           `tfsdk:"outbound_rule_security_group_ids"`
	OutboundRuleSecurityGroupNames types.List           `tfsdk:"outbound_rule_security_group_names"`
	OutboundRuleToPortRanges       types.List           `tfsdk:"outbound_rule_to_port_ranges"`
	SecurityGroupIds               types.List           `tfsdk:"security_group_ids"`
	SecurityGroupNames             types.List           `tfsdk:"security_group_names"`
	TagKeys                        types.List           `tfsdk:"tag_keys"`
	TagValues                      types.List           `tfsdk:"tag_values"`
	Tags                           types.List           `tfsdk:"tags"`
	VpcIds                         types.List           `tfsdk:"vpc_ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &securityGroupsDataSource{}
)

func (d *securityGroupsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.provider = provider
}

func NewSecurityGroupsDataSource() datasource.DataSource {
	return &securityGroupsDataSource{}
}

type securityGroupsDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *securityGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_groups"
}

// Schema defines the schema for the data source.
func (d *securityGroupsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SecurityGroupDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *securityGroupsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan SecurityGroupsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := d.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	params := SecurityGroupsFromTfToAPIReadParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.ReadSecurityGroupsResponse, error) {
		return numspotClient.ReadSecurityGroupsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Security Groups list")
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, SecurityGroupsFromHttpToTfDatasource, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
