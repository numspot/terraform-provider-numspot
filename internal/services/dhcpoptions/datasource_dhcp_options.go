package dhcpoptions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/dhcpoptions/datasource_dhcp_options"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &dhcpoptionsDataSource{}

// dhcpoptionsDataSource is the Terraform data source implementation for numspot_dhcp_options.
type dhcpoptionsDataSource struct {
	services.BaseDataSource
}

// NewDHCPOptionsDataSource returns a new dhcpoptionsDataSource instance.
func NewDHCPOptionsDataSource() datasource.DataSource {
	return &dhcpoptionsDataSource{}
}

// Metadata sets the data source type name.
func (d *dhcpoptionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_options"
}

// Schema defines the schema for the data source.
func (d *dhcpoptionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_dhcp_options.DhcpOptionsDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *dhcpoptionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_dhcp_options.DhcpOptionsModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadDHCPOptions(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadDHCPOptions(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read dhcp_options", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.DhcpOptionsSet) (datasource_dhcp_options.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	tagsList := types.ListNull(datasource_dhcp_options.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_dhcp_options.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_dhcp_options.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_dhcp_options.NewItemsValue(datasource_dhcp_options.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"default":             types.BoolValue(*item.Default),
		"domain_name":         types.StringValue(utils.ConvertStringPtrToString(item.DomainName)),
		"domain_name_servers": utils.FromStringListPointerToTfStringList(ctx, item.DomainNameServers, &diags),
		"id":                  types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"log_servers":         utils.FromStringListPointerToTfStringList(ctx, item.LogServers, &diags),
		"ntp_servers":         utils.FromStringListPointerToTfStringList(ctx, item.NtpServers, &diags),
		"tags":                tagsList,
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_dhcp_options.TagsValue, diag.Diagnostics) {
	return datasource_dhcp_options.NewTagsValue(datasource_dhcp_options.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func deserializeReadDHCPOptions(
	ctx context.Context,
	tf *datasource_dhcp_options.DhcpOptionsModel,
	diags *diag.Diagnostics,
) *api.ReadDhcpOptionsParams {
	return &api.ReadDhcpOptionsParams{
		Default:           tf.Default.ValueBoolPointer(),
		DomainNameServers: utils.ConvertTfListToArrayOfString(ctx, tf.DomainNameServers, diags),
		DomainNames:       utils.ConvertTfListToArrayOfString(ctx, tf.DomainNames, diags),
		LogServers:        utils.ConvertTfListToArrayOfString(ctx, tf.LogServers, diags),
		NtpServers:        utils.ConvertTfListToArrayOfString(ctx, tf.NtpServers, diags),
		TagKeys:           utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:         utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:              utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:               utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
