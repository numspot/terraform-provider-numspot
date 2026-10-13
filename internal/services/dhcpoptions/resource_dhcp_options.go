package dhcpoptions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/dhcpoptions/resource_dhcp_options"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &dhcpOptionsResource{}
	_ resource.ResourceWithConfigure   = &dhcpOptionsResource{}
	_ resource.ResourceWithImportState = &dhcpOptionsResource{}
)

// dhcpOptionsResource is the Terraform resource implementation for numspot_dhcp_options.
type dhcpOptionsResource struct {
	services.BaseResource
}

// NewDhcpOptionsResource returns a new dhcpOptionsResource instance.
func NewDhcpOptionsResource() resource.Resource {
	return &dhcpOptionsResource{}
}

// Metadata sets the resource type name.
func (r *dhcpOptionsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_dhcp_options"
}

// Schema defines the schema for the resource.
func (r *dhcpOptionsResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_dhcp_options.DhcpOptionsResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *dhcpOptionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_dhcp_options.DhcpOptionsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := dhcpoptionsTags(ctx, plan.Tags)
	numSpotDhcpOptions, err := core.CreateDHCPOptions(
		ctx,
		r.Provider,
		deserializeCreateDHCPOptionsRequest(ctx, &plan, &response.Diagnostics),
		tagsValue,
	)
	if err != nil {
		response.Diagnostics.AddError("unable to create dhcp_options", err.Error())
		return
	}

	state := serializeDHCPOptions(ctx, numSpotDhcpOptions, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *dhcpOptionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_dhcp_options.DhcpOptionsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotDhcpOptions, err := core.ReadDHCPOption(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read dhcp_options", err.Error())
		return
	}

	newState := serializeDHCPOptions(ctx, numSpotDhcpOptions, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *dhcpOptionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err                error
		state, plan        resource_dhcp_options.DhcpOptionsModel
		numSpotDhcpOptions *api.DhcpOptionsSet
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	dhcpoptionsID := state.Id.ValueString()
	planTags := dhcpoptionsTags(ctx, plan.Tags)
	stateTags := dhcpoptionsTags(ctx, state.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotDhcpOptions, err = core.UpdateDHCPOptionsTags(ctx, r.Provider, dhcpoptionsID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update dhcp_options tags", err.Error())
			return
		}
	}

	newState := serializeDHCPOptions(ctx, numSpotDhcpOptions, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *dhcpOptionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_dhcp_options.DhcpOptionsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteDHCPOptions(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete dhcp_options", err.Error())
		return
	}
}

func serializeDHCPOptions(ctx context.Context, http *api.DhcpOptionsSet, diags *diag.Diagnostics) resource_dhcp_options.DhcpOptionsModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil DhcpOptions response", "the API returned an empty DhcpOptions")
		return resource_dhcp_options.DhcpOptionsModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_dhcp_options.DhcpOptionsModel{
		Default:           types.BoolPointerValue(http.Default),
		DomainName:        types.StringPointerValue(http.DomainName),
		DomainNameServers: utils.FromStringListPointerToTfStringList(ctx, http.DomainNameServers, diags),
		Id:                types.StringPointerValue(http.Id),
		LogServers:        utils.FromStringListPointerToTfStringList(ctx, http.LogServers, diags),
		NtpServers:        utils.FromStringListPointerToTfStringList(ctx, http.NtpServers, diags),
		Tags:              tagsTf,
	}
}

func deserializeCreateDHCPOptionsRequest(
	ctx context.Context,
	tf *resource_dhcp_options.DhcpOptionsModel,
	diags *diag.Diagnostics,
) *api.CreateDhcpOptionsJSONRequestBody {
	return &api.CreateDhcpOptionsJSONRequestBody{
		DomainName:        tf.DomainName.ValueStringPointer(),
		DomainNameServers: utils.TfStringListToStringPtrList(ctx, tf.DomainNameServers, diags),
		LogServers:        utils.TfStringListToStringPtrList(ctx, tf.LogServers, diags),
		NtpServers:        utils.TfStringListToStringPtrList(ctx, tf.NtpServers, diags),
	}
}

func dhcpoptionsTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_dhcp_options.TagsValue, 0, len(tagsSet.Elements()))
	tagsSet.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}
