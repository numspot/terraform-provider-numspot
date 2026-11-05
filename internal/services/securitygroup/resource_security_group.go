package securitygroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/securitygroup/resource_security_group"
	"github.com/numspot/terraform-provider-numspot/internal/services/tags"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &securityGroupResource{}
	_ resource.ResourceWithConfigure   = &securityGroupResource{}
	_ resource.ResourceWithImportState = &securityGroupResource{}
)

// securityGroupResource is the Terraform resource implementation for numspot_security_group.
type securityGroupResource struct {
	services.BaseResource
}

// NewSecurityGroupResource returns a new securityGroupResource instance.
func NewSecurityGroupResource() resource.Resource {
	return &securityGroupResource{}
}

// Metadata sets the resource type name.
func (r *securityGroupResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_security_group"
}

// Schema defines the schema for the resource.
func (r *securityGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_security_group.SecurityGroupResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *securityGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_security_group.SecurityGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := securitygroupTags(ctx, plan.Tags)
	numspotSecurityGroup, err := core.CreateSecurityGroup(ctx, r.Provider, deserializeCreateSecurityGroupRequest(&plan), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create security_group", err.Error())
		return
	}

	state := serializeSecurityGroup(ctx, numspotSecurityGroup, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *securityGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_security_group.SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotSecurityGroup, err := core.ReadSecurityGroup(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read security_group", err.Error())
		return
	}

	newState := serializeSecurityGroup(ctx, numspotSecurityGroup, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *securityGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err                  error
		state, plan          resource_security_group.SecurityGroupModel
		numspotSecurityGroup *api.SecurityGroup
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	securitygroupID := state.Id.ValueString()
	planTags := securitygroupTags(ctx, plan.Tags)
	stateTags := securitygroupTags(ctx, state.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numspotSecurityGroup, err = core.UpdateSecurityGroupTags(ctx, r.Provider, securitygroupID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update security_group tags", err.Error())
			return
		}
	}

	newState := serializeSecurityGroup(ctx, numspotSecurityGroup, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *securityGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_security_group.SecurityGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteSecurityGroup(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete security_group", err.Error())
		return
	}
}

func serializeSecurityGroup(ctx context.Context, http *api.SecurityGroup, diags *diag.Diagnostics) resource_security_group.SecurityGroupModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil SecurityGroup response", "the API returned an empty SecurityGroup")
		return resource_security_group.SecurityGroupModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_security_group.SecurityGroupModel{
		Description: types.StringPointerValue(http.Description),
		Id:          types.StringPointerValue(http.Id),
		Name:        types.StringPointerValue(http.Name),
		VpcId:       types.StringPointerValue(http.VpcId),
		Tags:        tagsTf,
	}
}

func deserializeCreateSecurityGroupRequest(tf *resource_security_group.SecurityGroupModel) *api.CreateSecurityGroupJSONRequestBody {
	return &api.CreateSecurityGroupJSONRequestBody{
		Description: tf.Description.ValueString(),
		Name:        tf.Name.ValueString(),
		VpcId:       tf.VpcId.ValueString(),
	}
}

func securitygroupTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_security_group.TagsValue, 0, len(tagsSet.Elements()))
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
