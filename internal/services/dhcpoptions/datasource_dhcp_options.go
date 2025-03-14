package dhcpoptions

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/core"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
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
	resp.Schema = DhcpOptionsDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *dhcpOptionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan DHCPOptionsDataSourceModel

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

	dhcpOptionItems := serializeDHCPOptions(ctx, dhcpOptions, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = dhcpOptionItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func deserializeReadDHCPOptions(ctx context.Context, tf DHCPOptionsDataSourceModel, diags *diag.Diagnostics) numspot.ReadDhcpOptionsParams {
	ids := utils.TfStringListToStringPtrList(ctx, tf.IDs, diags)
	domainNames := utils.TfStringListToStringPtrList(ctx, tf.DomainNames, diags)
	dnsServers := utils.TfStringListToStringPtrList(ctx, tf.DomainNameServers, diags)
	logServers := utils.TfStringListToStringPtrList(ctx, tf.LogServers, diags)
	ntpServers := utils.TfStringListToStringPtrList(ctx, tf.NTPServers, diags)
	tagKeys := utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags)
	tagValues := utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags)
	numSpotTags := utils.TfStringListToStringPtrList(ctx, tf.Tags, diags)

	return numspot.ReadDhcpOptionsParams{
		Default:           tf.Default.ValueBoolPointer(),
		DomainNameServers: dnsServers,
		DomainNames:       domainNames,
		LogServers:        logServers,
		NtpServers:        ntpServers,
		TagKeys:           tagKeys,
		TagValues:         tagValues,
		Tags:              numSpotTags,
		Ids:               ids,
	}
}

func serializeDHCPOptions(ctx context.Context, dhcpOptions *[]numspot.DhcpOptionsSet, diags *diag.Diagnostics) []DhcpOptionsModel {
	return utils.FromHttpGenericListToTfList(ctx, dhcpOptions, func(ctx context.Context, dhcpOption *numspot.DhcpOptionsSet, diags *diag.Diagnostics) *DhcpOptionsModel {
		var tagsList types.List
		dnsServers := utils.FromStringListPointerToTfStringList(ctx, dhcpOption.DomainNameServers, diags)
		if diags.HasError() {
			return nil
		}
		logServers := utils.FromStringListPointerToTfStringList(ctx, dhcpOption.LogServers, diags)
		if diags.HasError() {
			return nil
		}
		ntpServers := utils.FromStringListPointerToTfStringList(ctx, dhcpOption.NtpServers, diags)
		if diags.HasError() {
			return nil
		}
		if dhcpOption.Tags != nil {
			tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *dhcpOption.Tags, diags)
			if diags.HasError() {
				return nil
			}
		}
		return &DhcpOptionsModel{
			Default:           types.BoolPointerValue(dhcpOption.Default),
			DomainName:        types.StringPointerValue(dhcpOption.DomainName),
			DomainNameServers: dnsServers,
			Id:                types.StringPointerValue(dhcpOption.Id),
			LogServers:        logServers,
			NtpServers:        ntpServers,
			Tags:              tagsList,
		}
	}, diags)
}
