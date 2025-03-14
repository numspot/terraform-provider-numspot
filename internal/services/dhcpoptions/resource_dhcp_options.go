package dhcpoptions

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/core"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                     = &Resource{}
	_ resource.ResourceWithConfigure        = &Resource{}
	_ resource.ResourceWithImportState      = &Resource{}
	_ resource.ResourceWithConfigValidators = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewDhcpOptionsResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	r.provider = provider
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_dhcp_options"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = DhcpOptionsResourceSchema(ctx)
}

func (r *Resource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.Root("domain_name").Expression(),
			path.Root("domain_name_servers").Expression(),
			path.Root("log_servers").Expression(),
			path.Root("ntp_servers").Expression(),
		),
	}
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan DhcpOptionsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)

	numSpotDHCPOptions, err := core.CreateDHCPOptions(ctx, r.provider, deserializeDHCPOption(ctx, plan), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create dhcp options", err.Error())
		return
	}

	state := serializeNumSpotDHCPOption(ctx, numSpotDHCPOptions, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state DhcpOptionsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	dhcpOptionsID := state.Id.ValueString()

	dhcpOptions, err := core.ReadDHCPOption(ctx, r.provider, dhcpOptionsID)
	if err != nil {
		response.Diagnostics.AddError("unable to read dhcp options", err.Error())
		return
	}

	newState := serializeNumSpotDHCPOption(ctx, dhcpOptions, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err                error
		numSpotDHCPOptions *numspot.DhcpOptionsSet
		state, plan        DhcpOptionsModel
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	dhcpOptionsID := state.Id.ValueString()
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotDHCPOptions, err = core.UpdateDHCPOptionsTags(ctx, r.provider, dhcpOptionsID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update dhcp options tags", err.Error())
			return
		}
	}

	newState := serializeNumSpotDHCPOption(ctx, numSpotDHCPOptions, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state DhcpOptionsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	dhcpOptionsID := state.Id.ValueString()

	if err := core.DeleteDHCPOptions(ctx, r.provider, dhcpOptionsID); err != nil {
		response.Diagnostics.AddError("unable to delete dhcp options", err.Error())
		return
	}
}

func deserializeDHCPOption(ctx context.Context, tf DhcpOptionsModel) numspot.CreateDhcpOptionsJSONRequestBody {
	var domainNameServers, logServers, ntpServers []string

	domainNameServers = make([]string, 0, len(tf.DomainNameServers.Elements()))
	tf.DomainNameServers.ElementsAs(ctx, &domainNameServers, false)

	logServers = make([]string, 0, len(tf.LogServers.Elements()))
	tf.LogServers.ElementsAs(ctx, &logServers, false)

	ntpServers = make([]string, 0, len(tf.NtpServers.Elements()))
	tf.NtpServers.ElementsAs(ctx, &ntpServers, false)

	return numspot.CreateDhcpOptionsJSONRequestBody{
		DomainName:        tf.DomainName.ValueStringPointer(),
		DomainNameServers: &domainNameServers,
		LogServers:        &logServers,
		NtpServers:        &ntpServers,
	}
}

func serializeNumSpotDHCPOption(ctx context.Context, http *numspot.DhcpOptionsSet, diags *diag.Diagnostics) DhcpOptionsModel {
	var domainNameServersTf, logServersTf, ntpServersTf, tagsTf types.List

	if http.DomainNameServers != nil {
		domainNameServersTf = utils.StringListToTfListValue(ctx, *http.DomainNameServers, diags)
		if diags.HasError() {
			return DhcpOptionsModel{}
		}
	}

	if http.LogServers != nil {
		logServersTf = utils.StringListToTfListValue(ctx, *http.LogServers, diags)
		if diags.HasError() {
			return DhcpOptionsModel{}
		}
	}

	if http.NtpServers != nil {
		ntpServersTf = utils.StringListToTfListValue(ctx, *http.NtpServers, diags)
		if diags.HasError() {
			return DhcpOptionsModel{}
		}
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return DhcpOptionsModel{}
		}
	}

	return DhcpOptionsModel{
		Default:           types.BoolPointerValue(http.Default),
		DomainName:        types.StringPointerValue(http.DomainName),
		Id:                types.StringPointerValue(http.Id),
		DomainNameServers: domainNameServersTf,
		LogServers:        logServersTf,
		NtpServers:        ntpServersTf,
		Tags:              tagsTf,
	}
}
