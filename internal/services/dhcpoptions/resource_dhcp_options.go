package dhcpoptions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/dhcpoptions/resource_dhcp_options"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                     = &dhcpOptionsResource{}
	_ resource.ResourceWithConfigure        = &dhcpOptionsResource{}
	_ resource.ResourceWithImportState      = &dhcpOptionsResource{}
	_ resource.ResourceWithConfigValidators = &dhcpOptionsResource{}
)

type dhcpOptionsResource struct {
	provider *client.NumSpotSDK
}

func NewDhcpOptionsResource() resource.Resource {
	return &dhcpOptionsResource{}
}

func (r *dhcpOptionsResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *dhcpOptionsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *dhcpOptionsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_dhcp_options"
}

func (r *dhcpOptionsResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_dhcp_options.DhcpOptionsResourceSchema(ctx)
}

func (r *dhcpOptionsResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.Root("domain_name").Expression(),
			path.Root("domain_name_servers").Expression(),
			path.Root("log_servers").Expression(),
			path.Root("ntp_servers").Expression(),
		),
	}
}

func (r *dhcpOptionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_dhcp_options.DhcpOptionsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	apiTags := dhcpTags(ctx, plan.Tags)

	numSpotDHCPOptions, err := core.CreateDHCPOptions(ctx, r.provider, deserializeDHCPOption(ctx, plan), apiTags)
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

func (r *dhcpOptionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_dhcp_options.DhcpOptionsModel
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

func (r *dhcpOptionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err                error
		numSpotDHCPOptions *api.DhcpOptionsSet
		state, plan        resource_dhcp_options.DhcpOptionsModel
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
	stateTags := dhcpTags(ctx, state.Tags)
	planTags := dhcpTags(ctx, plan.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotDHCPOptions, err = core.UpdateDHCPOptionsTags(ctx, r.provider, dhcpOptionsID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update dhcp options tags", err.Error())
			return
		}

		newState := serializeNumSpotDHCPOption(ctx, numSpotDHCPOptions, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
	}
}

func (r *dhcpOptionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_dhcp_options.DhcpOptionsModel
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

func deserializeDHCPOption(ctx context.Context, tf resource_dhcp_options.DhcpOptionsModel) api.CreateDhcpOptionsJSONRequestBody {
	var domainNameServers, logServers, ntpServers []string

	domainNameServers = make([]string, 0, len(tf.DomainNameServers.Elements()))
	tf.DomainNameServers.ElementsAs(ctx, &domainNameServers, false)

	logServers = make([]string, 0, len(tf.LogServers.Elements()))
	tf.LogServers.ElementsAs(ctx, &logServers, false)

	ntpServers = make([]string, 0, len(tf.NtpServers.Elements()))
	tf.NtpServers.ElementsAs(ctx, &ntpServers, false)

	return api.CreateDhcpOptionsJSONRequestBody{
		DomainName:        tf.DomainName.ValueStringPointer(),
		DomainNameServers: &domainNameServers,
		LogServers:        &logServers,
		NtpServers:        &ntpServers,
	}
}

func serializeNumSpotDHCPOption(ctx context.Context, http *api.DhcpOptionsSet, diags *diag.Diagnostics) resource_dhcp_options.DhcpOptionsModel {
	var domainNameServersTf, logServersTf, ntpServersTf types.List
	var tagsTf types.Set

	if http.DomainNameServers != nil {
		domainNameServersTf = utils.StringListToTfListValue(ctx, *http.DomainNameServers, diags)
		if diags.HasError() {
			return resource_dhcp_options.DhcpOptionsModel{}
		}
	}

	if http.LogServers != nil {
		logServersTf = utils.StringListToTfListValue(ctx, *http.LogServers, diags)
		if diags.HasError() {
			return resource_dhcp_options.DhcpOptionsModel{}
		}
	}

	if http.NtpServers != nil {
		ntpServersTf = utils.StringListToTfListValue(ctx, *http.NtpServers, diags)
		if diags.HasError() {
			return resource_dhcp_options.DhcpOptionsModel{}
		}
	}

	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return resource_dhcp_options.DhcpOptionsModel{}
		}
	}

	return resource_dhcp_options.DhcpOptionsModel{
		Default:           types.BoolPointerValue(http.Default),
		DomainName:        types.StringPointerValue(http.DomainName),
		Id:                types.StringPointerValue(http.Id),
		DomainNameServers: domainNameServersTf,
		LogServers:        logServersTf,
		NtpServers:        ntpServersTf,
		Tags:              tagsTf,
	}
}

func dhcpTags(ctx context.Context, tags types.Set) []api.ResourceTag {
	tfTags := make([]resource_dhcp_options.TagsValue, 0, len(tags.Elements()))
	tags.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}
