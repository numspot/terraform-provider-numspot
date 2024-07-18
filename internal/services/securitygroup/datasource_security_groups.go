package securitygroup

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type SecurityGroupsDataSourceModel struct {
	Items                         []SecurityGroupModel `tfsdk:"items"`
	Descriptions                  types.List           `tfsdk:"descriptions"`
	InboundRulesAccountIds        types.List           `tfsdk:"inbound_rules_account_ids"`
	InboundRulesFromPortRanges    types.List           `tfsdk:"inbound_rules_from_port_ranges"`
	InboundRulesRuleProtocols     types.List           `tfsdk:"inbound_rules_rule_protocols"`
	InboundRulesIpRanges          types.List           `tfsdk:"inbound_rules_ip_ranges"`
	InboundRulesSecurityGroupIds  types.List           `tfsdk:"inbound_rules_security_group_ids"`
	InboundRulesToPortRanges      types.List           `tfsdk:"inbound_rules_to_port_ranges"`
	Ids                           types.List           `tfsdk:"ids"`
	Names                         types.List           `tfsdk:"names"`
	OutboundRulesFromPortRanges   types.List           `tfsdk:"outbound_rules_from_port_ranges"`
	OutboundRulesRuleProtocols    types.List           `tfsdk:"outbound_rules_rule_protocols"`
	OutboundRulesIpRanges         types.List           `tfsdk:"outbound_rules_ip_ranges"`
	OutboundRulesAccountIds       types.List           `tfsdk:"outbound_rules_account_ids"`
	OutboundRulesSecurityGroupIds types.List           `tfsdk:"outbound_rules_security_group_ids"`
	OutboundRulesToPortRanges     types.List           `tfsdk:"outbound_rules_to_port_ranges"`
	VpcIds                        types.List           `tfsdk:"vpc_ids"`
	TagKeys                       types.List           `tfsdk:"tag_keys"`
	TagValues                     types.List           `tfsdk:"tag_values"`
	Tags                          types.List           `tfsdk:"tags"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &securityGroupsDataSource{}
)

func (d *securityGroupsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(services.IProvider)
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
	provider services.IProvider
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

	params := SecurityGroupsFromTfToAPIReadParams(ctx, plan)
	res := utils2.ExecuteRequest(func() (*numspot.ReadSecurityGroupsResponse, error) {
		return d.provider.GetNumspotClient().ReadSecurityGroupsWithResponse(ctx, d.provider.GetSpaceID(), &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Security Groups list")
	}

	objectItems, diags := utils2.FromHttpGenericListToTfList(ctx, res.JSON200.Items, SecurityGroupsFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
