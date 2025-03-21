package securitygroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/securitygroup/datasource_security_group"
	"terraform-provider-numspot/internal/utils"
)

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
	resp.Schema = datasource_security_group.SecurityGroupDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *securityGroupsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_security_group.SecurityGroupModel
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

	objectItems := serializeSecurityGroupsDatasource(ctx, res, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeSecurityGroupsDatasourceParams(ctx context.Context, tf datasource_security_group.SecurityGroupModel, diags *diag.Diagnostics) api.ReadSecurityGroupsParams {
	return api.ReadSecurityGroupsParams{
		Descriptions:                 utils.ConvertTfListToArrayOfString(ctx, tf.Descriptions, diags),
		InboundRuleFromPortRanges:    utils.ConvertTfListToArrayOfInt(ctx, tf.InboundRuleFromPortRanges, diags),
		InboundRuleProtocols:         utils.ConvertTfListToArrayOfString(ctx, tf.InboundRuleProtocols, diags),
		InboundRuleIpRanges:          utils.ConvertTfListToArrayOfString(ctx, tf.InboundRuleIpRanges, diags),
		InboundRuleSecurityGroupIds:  utils.ConvertTfListToArrayOfString(ctx, tf.InboundRuleSecurityGroupIds, diags),
		InboundRuleToPortRanges:      utils.ConvertTfListToArrayOfInt(ctx, tf.InboundRuleToPortRanges, diags),
		SecurityGroupIds:             utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:           utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupNames, diags),
		OutboundRuleFromPortRanges:   utils.ConvertTfListToArrayOfInt(ctx, tf.OutboundRuleFromPortRanges, diags),
		OutboundRuleProtocols:        utils.ConvertTfListToArrayOfString(ctx, tf.OutboundRuleProtocols, diags),
		OutboundRuleIpRanges:         utils.ConvertTfListToArrayOfString(ctx, tf.OutboundRuleIpRanges, diags),
		OutboundRuleSecurityGroupIds: utils.ConvertTfListToArrayOfString(ctx, tf.OutboundRuleSecurityGroupIds, diags),
		OutboundRuleToPortRanges:     utils.ConvertTfListToArrayOfInt(ctx, tf.OutboundRuleToPortRanges, diags),
		VpcIds:                       utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		TagKeys:                      utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:                    utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:                         utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
	}
}

func serializeSecurityGroupsDatasource(ctx context.Context, securityGroups *[]api.SecurityGroup, diags *diag.Diagnostics) datasource_security_group.SecurityGroupModel {
	var securityGroupsList types.List
	tagsList := types.List{}
	inboundRulesList := types.List{}
	outboundRulesList := types.List{}

	var serializeDiags diag.Diagnostics

	if len(*securityGroups) != 0 {
		ll := len(*securityGroups)
		itemsValue := make([]datasource_security_group.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			if (*securityGroups)[i].Tags != nil {
				tagsList, serializeDiags = mappingSecurityGroupTags(ctx, securityGroups, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*securityGroups)[i].InboundRules != nil {
				inboundRulesList, serializeDiags = mappingInboundRules(ctx, securityGroups, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*securityGroups)[i].OutboundRules != nil {
				outboundRulesList, serializeDiags = mappingOutboundRules(ctx, securityGroups, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			itemsValue[i], serializeDiags = datasource_security_group.NewItemsValue(datasource_security_group.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"description":    types.StringValue(utils.ConvertStringPtrToString((*securityGroups)[i].Description)),
				"id":             types.StringValue(utils.ConvertStringPtrToString((*securityGroups)[i].Id)),
				"inbound_rules":  inboundRulesList,
				"name":           types.StringValue(utils.ConvertStringPtrToString((*securityGroups)[i].Name)),
				"outbound_rules": outboundRulesList,
				"tags":           tagsList,
				"vpc_id":         types.StringValue(utils.ConvertStringPtrToString((*securityGroups)[i].VpcId)),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		securityGroupsList, serializeDiags = types.ListValueFrom(ctx, new(datasource_security_group.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		securityGroupsList = types.ListNull(new(datasource_security_group.ItemsValue).Type(ctx))
	}

	return datasource_security_group.SecurityGroupModel{
		Items: securityGroupsList,
	}
}

func mappingSecurityGroupTags(ctx context.Context, securityGroups *[]api.SecurityGroup, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	lt := len(*(*securityGroups)[i].Tags)
	elementValue := make([]datasource_security_group.TagsValue, lt)
	for y, tag := range *(*securityGroups)[i].Tags {
		elementValue[y], *diags = datasource_security_group.NewTagsValue(datasource_security_group.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"key":   types.StringValue(tag.Key),
			"value": types.StringValue(tag.Value),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_security_group.TagsValue).Type(ctx), elementValue)
}

func mappingInboundRules(ctx context.Context, securityGroups *[]api.SecurityGroup, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	var mappingDiags diag.Diagnostics
	var ipRangesList types.List
	var serviceIdsList types.List

	lt := len(*(*securityGroups)[i].InboundRules)
	elementValue := make([]datasource_security_group.InboundRulesValue, lt)
	for y, rule := range *(*securityGroups)[i].InboundRules {

		ipRangesList, mappingDiags = types.ListValueFrom(ctx, types.StringType, rule.IpRanges)
		diags.Append(mappingDiags...)

		serviceIdsList, mappingDiags = types.ListValueFrom(ctx, types.StringType, rule.ServiceIds)
		diags.Append(mappingDiags...)

		if rule.ServiceIds == nil {
			serviceIdsList, mappingDiags = types.ListValueFrom(ctx, types.StringType, []string{})
			diags.Append(mappingDiags...)
		}

		elementValue[y], *diags = datasource_security_group.NewInboundRulesValue(datasource_security_group.InboundRulesValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"from_port_range":                 types.Int64Value(utils.ConvertIntPtrToInt64(rule.FromPortRange)),
			"inbound_security_groups_members": types.ListNull(datasource_security_group.InboundSecurityGroupsMembersValue{}.Type(ctx)),
			"ip_protocol":                     types.StringValue(utils.ConvertStringPtrToString(rule.IpProtocol)),
			"ip_ranges":                       ipRangesList,
			"service_ids":                     serviceIdsList,
			"to_port_range":                   types.Int64Value(utils.ConvertIntPtrToInt64(rule.ToPortRange)),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_security_group.InboundRulesValue).Type(ctx), elementValue)
}

func mappingOutboundRules(ctx context.Context, securityGroups *[]api.SecurityGroup, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	var mappingDiags diag.Diagnostics
	var ipRangesList types.List
	var serviceIdsList types.List

	lt := len(*(*securityGroups)[i].OutboundRules)
	elementValue := make([]datasource_security_group.OutboundRulesValue, lt)
	for y, rule := range *(*securityGroups)[i].OutboundRules {

		ipRangesList, mappingDiags = types.ListValueFrom(ctx, types.StringType, rule.IpRanges)
		diags.Append(mappingDiags...)

		serviceIdsList, mappingDiags = types.ListValueFrom(ctx, types.StringType, rule.ServiceIds)
		diags.Append(mappingDiags...)

		if rule.ServiceIds == nil {
			serviceIdsList, mappingDiags = types.ListValueFrom(ctx, types.StringType, []string{})
			diags.Append(mappingDiags...)
		}

		elementValue[y], *diags = datasource_security_group.NewOutboundRulesValue(datasource_security_group.OutboundRulesValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"from_port_range":                  types.Int64Value(utils.ConvertIntPtrToInt64(rule.FromPortRange)),
			"ip_protocol":                      types.StringValue(utils.ConvertStringPtrToString(rule.IpProtocol)),
			"ip_ranges":                        ipRangesList,
			"outbound_security_groups_members": types.ListNull(datasource_security_group.OutboundSecurityGroupsMembersValue{}.Type(ctx)),
			"service_ids":                      serviceIdsList,
			"to_port_range":                    types.Int64Value(utils.ConvertIntPtrToInt64(rule.ToPortRange)),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_security_group.OutboundRulesValue).Type(ctx), elementValue)
}
