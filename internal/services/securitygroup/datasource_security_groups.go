package securitygroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/core"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
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

	params := deserializeSecurityGroupsDatasourceParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := core.ReadSecurityGroups(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("failed to read security groups", err.Error())
		return
	}

	if res == nil {
		response.Diagnostics.AddError("failed to read security groups", "got empty Security Groups list")
		return
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, res, serializeSecurityGroupsDatasource, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeSecurityGroupsDatasourceParams(ctx context.Context, tf SecurityGroupsDataSourceModel, diags *diag.Diagnostics) numspot.ReadSecurityGroupsParams {
	return numspot.ReadSecurityGroupsParams{
		Descriptions:                 utils.TfStringListToStringPtrList(ctx, tf.Descriptions, diags),
		InboundRuleFromPortRanges:    utils.TFInt64ListToIntListPointer(ctx, tf.InboundRuleFromPortRanges, diags),
		InboundRuleProtocols:         utils.TfStringListToStringPtrList(ctx, tf.InboundRuleProtocols, diags),
		InboundRuleIpRanges:          utils.TfStringListToStringPtrList(ctx, tf.InboundRuleIpRanges, diags),
		InboundRuleSecurityGroupIds:  utils.TfStringListToStringPtrList(ctx, tf.InboundRuleSecurityGroupIds, diags),
		InboundRuleToPortRanges:      utils.TFInt64ListToIntListPointer(ctx, tf.InboundRuleToPortRanges, diags),
		SecurityGroupIds:             utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:           utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames, diags),
		OutboundRuleFromPortRanges:   utils.TFInt64ListToIntListPointer(ctx, tf.OutboundRuleFromPortRanges, diags),
		OutboundRuleProtocols:        utils.TfStringListToStringPtrList(ctx, tf.OutboundRuleProtocols, diags),
		OutboundRuleIpRanges:         utils.TfStringListToStringPtrList(ctx, tf.OutboundRuleIpRanges, diags),
		OutboundRuleSecurityGroupIds: utils.TfStringListToStringPtrList(ctx, tf.OutboundRuleSecurityGroupIds, diags),
		OutboundRuleToPortRanges:     utils.TFInt64ListToIntListPointer(ctx, tf.OutboundRuleToPortRanges, diags),
		VpcIds:                       utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		TagKeys:                      utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:                    utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:                         utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
	}
}

func serializeSecurityGroupsDatasource(ctx context.Context, http *numspot.SecurityGroup, diags *diag.Diagnostics) *SecurityGroupModel {
	var (
		inboundRules  = types.SetNull(InboundRulesValue{}.Type(ctx))
		outboundRules = types.SetNull(OutboundRulesValue{}.Type(ctx))
		tagsList      types.List
	)

	if http.InboundRules != nil {
		inboundRules = utils.GenericSetToTfSetValue(
			ctx,
			serializeInboundRuleDatasource,
			*http.InboundRules,
			diags,
		)
		if diags.HasError() {
			return nil
		}
	}

	if http.OutboundRules != nil {
		outboundRules = utils.GenericSetToTfSetValue(
			ctx,
			serializeOutboundRuleDatasource,
			*http.OutboundRules,
			diags,
		)
		if diags.HasError() {
			return nil
		}
	}

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &SecurityGroupModel{
		Id:            types.StringPointerValue(http.Id),
		Tags:          tagsList,
		Description:   types.StringPointerValue(http.Description),
		InboundRules:  inboundRules,
		Name:          types.StringPointerValue(http.Name),
		OutboundRules: outboundRules,
		VpcId:         types.StringPointerValue(http.VpcId),
	}
}

func serializeInboundRuleDatasource(ctx context.Context, rules numspot.SecurityGroupRule, diags *diag.Diagnostics) InboundRulesValue {
	ipRanges, diagnostics := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	diags.Append(diagnostics...)
	serviceIds, diagnostics := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	diags.Append(diagnostics...)
	if rules.ServiceIds == nil {
		serviceIds, diagnostics = types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(diagnostics...)
	}

	value, diagnostics := NewInboundRulesValue(
		InboundRulesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"from_port_range":         utils.FromIntPtrToTfInt64(rules.FromPortRange),
			"to_port_range":           utils.FromIntPtrToTfInt64(rules.ToPortRange),
			"ip_protocol":             types.StringPointerValue(rules.IpProtocol),
			"ip_ranges":               ipRanges,
			"security_groups_members": types.ListNull(SecurityGroupsMembersValue{}.Type(ctx)),
			"service_ids":             serviceIds,
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeOutboundRuleDatasource(ctx context.Context, rules numspot.SecurityGroupRule, diags *diag.Diagnostics) OutboundRulesValue {
	ipRanges, diagnostics := types.ListValueFrom(ctx, types.StringType, rules.IpRanges)
	diags.Append(diagnostics...)

	serviceIds, diagnostics := types.ListValueFrom(ctx, types.StringType, rules.ServiceIds)
	diags.Append(diagnostics...)

	if rules.ServiceIds == nil {
		serviceIds, diagnostics = types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(diagnostics...)
	}

	value, diagnostics := NewOutboundRulesValue(
		OutboundRulesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"from_port_range":         utils.FromIntPtrToTfInt64(rules.FromPortRange),
			"to_port_range":           utils.FromIntPtrToTfInt64(rules.ToPortRange),
			"ip_protocol":             types.StringPointerValue(rules.IpProtocol),
			"ip_ranges":               ipRanges,
			"security_groups_members": types.ListNull(SecurityGroupsMembersValue{}.Type(ctx)),
			"service_ids":             serviceIds,
		},
	)
	diags.Append(diagnostics...)
	return value
}
