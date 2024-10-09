package dhcpoptions

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type DHCPOptionsDataSourceModel struct {
	Items             []DhcpOptionsModel `tfsdk:"items"`
	IDs               types.List         `tfsdk:"ids"`
	Default           types.Bool         `tfsdk:"default"`
	DomainNameServers types.List         `tfsdk:"domain_name_servers"`
	DomainNames       types.List         `tfsdk:"domain_names"`
	LogServers        types.List         `tfsdk:"log_servers"`
	NTPServers        types.List         `tfsdk:"ntp_servers"`
	TagKeys           types.List         `tfsdk:"tag_keys"`
	TagValues         types.List         `tfsdk:"tag_values"`
	Tags              types.List         `tfsdk:"tags"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &dhcpOptionsDataSource{}
)

func (d *dhcpOptionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewDHCPOptionsDataSource() datasource.DataSource {
	return &dhcpOptionsDataSource{}
}

type dhcpOptionsDataSource struct {
	provider services.IProvider
}

// Metadata returns the data source type name.
func (d *dhcpOptionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_options"
}

// Schema defines the schema for the data source.
func (d *dhcpOptionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DhcpOptionsDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *dhcpOptionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan DHCPOptionsDataSourceModel
	var diags diag.Diagnostics
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := dhcpOptionsFromTfToAPIReadParams(ctx, plan, &diags)

	dhcpOptions, err := core.ReadDHCPOptions(ctx, d.provider, params)
	if err != nil {
		return
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, dhcpOptions.Items, dhcpOptionsFromHttpToTfDatasource, &diags)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func dhcpOptionsFromTfToAPIReadParams(ctx context.Context, tf DHCPOptionsDataSourceModel, diags *diag.Diagnostics) numspot.ReadDhcpOptionsParams {
	ids := utils.TfStringListToStringPtrList(ctx, tf.IDs, diags)
	domainNames := utils.TfStringListToStringPtrList(ctx, tf.DomainNames, diags)
	dnsServers := utils.TfStringListToStringPtrList(ctx, tf.DomainNameServers, diags)
	logServers := utils.TfStringListToStringPtrList(ctx, tf.LogServers, diags)
	ntpServers := utils.TfStringListToStringPtrList(ctx, tf.NTPServers, diags)
	tagKeys := utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags)
	tagValues := utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags)
	tags := utils.TfStringListToStringPtrList(ctx, tf.Tags, diags)

	return numspot.ReadDhcpOptionsParams{
		Default:           tf.Default.ValueBoolPointer(),
		DomainNameServers: dnsServers,
		DomainNames:       domainNames,
		LogServers:        logServers,
		NtpServers:        ntpServers,
		TagKeys:           tagKeys,
		TagValues:         tagValues,
		Tags:              tags,
		Ids:               ids,
	}
}

func dhcpOptionsFromHttpToTfDatasource(ctx context.Context, http *numspot.DhcpOptionsSet, diags *diag.Diagnostics) *DhcpOptionsModel {
	var tagsList types.List
	dnsServers := utils.FromStringListPointerToTfStringList(ctx, http.DomainNameServers, diags)
	if diags.HasError() {
		return nil
	}
	logServers := utils.FromStringListPointerToTfStringList(ctx, http.LogServers, diags)
	if diags.HasError() {
		return nil
	}
	ntpServers := utils.FromStringListPointerToTfStringList(ctx, http.NtpServers, diags)
	if diags.HasError() {
		return nil
	}
	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}
	return &DhcpOptionsModel{
		Default:           types.BoolPointerValue(http.Default),
		DomainName:        types.StringPointerValue(http.DomainName),
		DomainNameServers: dnsServers,
		Id:                types.StringPointerValue(http.Id),
		LogServers:        logServers,
		NtpServers:        ntpServers,
		Tags:              tagsList,
	}
}
