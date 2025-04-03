package dhcpoptions

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
	"terraform-provider-numspot/internal/services/dhcpoptions/datasource_dhcp_options"
	"terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &dhcpOptionsDataSource{}
)

func (d *dhcpOptionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewDHCPOptionsDataSource() datasource.DataSource {
	return &dhcpOptionsDataSource{}
}

type dhcpOptionsDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *dhcpOptionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_options"
}

// Schema defines the schema for the data source.
func (d *dhcpOptionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_dhcp_options.DhcpOptionsDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *dhcpOptionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_dhcp_options.DhcpOptionsModel

	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	dhcpOptionParams := deserializeReadDHCPOptions(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	dhcpOptions, err := core.ReadDHCPOptions(ctx, d.provider, dhcpOptionParams)
	if err != nil {
		response.Diagnostics.AddError("unable to read dhcp options", err.Error())
		return
	}

	dhcpOptionItems := utils.SerializeDatasourceItemsWithDiags(ctx, *dhcpOptions, &response.Diagnostics, mappingItemsValue)
	if response.Diagnostics.HasError() {
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, dhcpOptionItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func deserializeReadDHCPOptions(ctx context.Context, tf datasource_dhcp_options.DhcpOptionsModel, diags *diag.Diagnostics) api.ReadDhcpOptionsParams {
	return api.ReadDhcpOptionsParams{
		Default:           tf.Default.ValueBoolPointer(),
		DomainNameServers: utils.ConvertTfListToArrayOfString(ctx, tf.DomainNameServers, diags),
		DomainNames:       utils.ConvertTfListToArrayOfString(ctx, tf.DomainNames, diags),
		LogServers:        utils.ConvertTfListToArrayOfString(ctx, tf.LogServers, diags),
		NtpServers:        utils.ConvertTfListToArrayOfString(ctx, tf.NtpServers, diags),
		TagKeys:           utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:         utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:              utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:               utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
	}
}

func mappingItemsValue(ctx context.Context, dhcpOption api.DhcpOptionsSet, diags *diag.Diagnostics) (datasource_dhcp_options.ItemsValue, diag.Diagnostics) {
	var serializeDiags diag.Diagnostics
	domainNameServersList := types.List{}
	logServersList := types.List{}
	ntpServersList := types.List{}
	tagsList := types.ListNull(datasource_dhcp_options.ItemsValue{}.Type(ctx))

	if dhcpOption.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *dhcpOption.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_dhcp_options.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_dhcp_options.ItemsValue{}, serializeDiags
		}
	}

	if dhcpOption.DomainNameServers != nil {
		domainNameServersList, serializeDiags = types.ListValueFrom(ctx, types.StringType, dhcpOption.DomainNameServers)
		diags.Append(serializeDiags...)
	}

	if dhcpOption.LogServers != nil {
		logServersList, serializeDiags = types.ListValueFrom(ctx, types.StringType, dhcpOption.LogServers)
		diags.Append(serializeDiags...)
	}

	if dhcpOption.NtpServers != nil {
		ntpServersList, serializeDiags = types.ListValueFrom(ctx, types.StringType, dhcpOption.NtpServers)
		diags.Append(serializeDiags...)
	}

	return datasource_dhcp_options.NewItemsValue(datasource_dhcp_options.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"default":             types.BoolPointerValue(dhcpOption.Default),
		"domain_name":         types.StringValue(utils.ConvertStringPtrToString(dhcpOption.DomainName)),
		"domain_name_servers": domainNameServersList,
		"id":                  types.StringValue(utils.ConvertStringPtrToString(dhcpOption.Id)),
		"log_servers":         logServersList,
		"ntp_servers":         ntpServersList,
		"tags":                tagsList,
	})
}

func mappingTags(ctx context.Context, tag api.ResourceTag) (datasource_dhcp_options.TagsValue, diag.Diagnostics) {
	return datasource_dhcp_options.NewTagsValue(datasource_dhcp_options.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}
