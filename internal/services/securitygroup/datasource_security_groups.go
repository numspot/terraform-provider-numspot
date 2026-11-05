package securitygroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/securitygroup/datasource_security_group"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &securitygroupsDataSource{}

// securitygroupsDataSource is the Terraform data source implementation for numspot_security_groups.
type securitygroupsDataSource struct {
	services.BaseDataSource
}

// NewSecurityGroupsDataSource returns a new securitygroupsDataSource instance.
func NewSecurityGroupsDataSource() datasource.DataSource {
	return &securitygroupsDataSource{}
}

// Metadata sets the data source type name.
func (d *securitygroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_groups"
}

// Schema defines the schema for the data source.
func (d *securitygroupsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_security_group.SecurityGroupDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *securitygroupsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_security_group.SecurityGroupModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadSecurityGroups(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadSecurityGroups(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read security_groups", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.SecurityGroup) (datasource_security_group.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_security_group.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_security_group.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_security_group.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_security_group.NewItemsValue(datasource_security_group.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"description": types.StringValue(utils.ConvertStringPtrToString(item.Description)),
		"id":          types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"name":        types.StringValue(utils.ConvertStringPtrToString(item.Name)),
		"vpc_id":      types.StringValue(utils.ConvertStringPtrToString(item.VpcId)),
		"tags":        tagsList,
	})
	return value, valueDiags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_security_group.TagsValue, diag.Diagnostics) {
	return datasource_security_group.NewTagsValue(datasource_security_group.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func deserializeReadSecurityGroups(
	ctx context.Context,
	tf *datasource_security_group.SecurityGroupModel,
	diags *diag.Diagnostics,
) *api.ReadSecurityGroupsParams {
	return &api.ReadSecurityGroupsParams{
		Descriptions:                   utils.ConvertTfListToArrayOfString(ctx, tf.Descriptions, diags),
		InboundRuleFromPortRanges:      utils.ConvertTfListToArrayOfInt(ctx, tf.InboundRuleFromPortRanges, diags),
		InboundRuleIpRanges:            utils.ConvertTfListToArrayOfString(ctx, tf.InboundRuleIpRanges, diags),
		InboundRuleProtocols:           utils.ConvertTfListToArrayOfString(ctx, tf.InboundRuleProtocols, diags),
		InboundRuleSecurityGroupIds:    utils.ConvertTfListToArrayOfString(ctx, tf.InboundRuleSecurityGroupIds, diags),
		InboundRuleSecurityGroupNames:  utils.ConvertTfListToArrayOfString(ctx, tf.InboundRuleSecurityGroupNames, diags),
		InboundRuleToPortRanges:        utils.ConvertTfListToArrayOfInt(ctx, tf.InboundRuleToPortRanges, diags),
		OutboundRuleFromPortRanges:     utils.ConvertTfListToArrayOfInt(ctx, tf.OutboundRuleFromPortRanges, diags),
		OutboundRuleIpRanges:           utils.ConvertTfListToArrayOfString(ctx, tf.OutboundRuleIpRanges, diags),
		OutboundRuleProtocols:          utils.ConvertTfListToArrayOfString(ctx, tf.OutboundRuleProtocols, diags),
		OutboundRuleSecurityGroupIds:   utils.ConvertTfListToArrayOfString(ctx, tf.OutboundRuleSecurityGroupIds, diags),
		OutboundRuleSecurityGroupNames: utils.ConvertTfListToArrayOfString(ctx, tf.OutboundRuleSecurityGroupNames, diags),
		OutboundRuleToPortRanges:       utils.ConvertTfListToArrayOfInt(ctx, tf.OutboundRuleToPortRanges, diags),
		SecurityGroupIds:               utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:             utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupNames, diags),
		TagKeys:                        utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:                      utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:                           utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		VpcIds:                         utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
