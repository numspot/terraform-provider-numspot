package flexiblegpu

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/flexiblegpu/resource_flexible_gpu"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &flexibleGpuResource{}
	_ resource.ResourceWithConfigure   = &flexibleGpuResource{}
	_ resource.ResourceWithImportState = &flexibleGpuResource{}
)

// flexibleGpuResource is the Terraform resource implementation for numspot_flexible_gpu.
type flexibleGpuResource struct {
	services.BaseResource
}

// NewFlexibleGpuResource returns a new flexibleGpuResource instance.
func NewFlexibleGpuResource() resource.Resource {
	return &flexibleGpuResource{}
}

// Metadata sets the resource type name.
func (r *flexibleGpuResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_flexible_gpu"
}

// Schema defines the schema for the resource.
func (r *flexibleGpuResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_flexible_gpu.FlexibleGpuResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *flexibleGpuResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numspotFlexibleGpu, err := core.CreateFlexibleGPU(ctx, r.Provider, deserializeCreateFlexibleGPURequest(&plan), vmIDFromModel(&plan))
	if err != nil {
		response.Diagnostics.AddError("unable to create flexible_gpu", err.Error())
		return
	}

	state := serializeFlexibleGPU(numspotFlexibleGpu)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *flexibleGpuResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotFlexibleGpu, err := core.ReadFlexibleGPU(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read flexible_gpu", err.Error())
		return
	}

	newState := serializeFlexibleGPU(numspotFlexibleGpu)

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *flexibleGpuResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err                error
		state, plan        resource_flexible_gpu.FlexibleGpuModel
		numspotFlexibleGpu *api.FlexibleGpu
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.DeleteOnVmDeletion.Equal(state.DeleteOnVmDeletion) || !plan.VmId.Equal(state.VmId) {
		numspotFlexibleGpu, err = core.UpdateFlexibleGPU(
			ctx,
			r.Provider,
			state.Id.ValueString(),
			vmIDFromModel(&state),
			vmIDFromModel(&plan),
			!plan.DeleteOnVmDeletion.Equal(state.DeleteOnVmDeletion),
			deserializeUpdateFlexibleGpu(&plan),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update flexible_gpu", err.Error())
			return
		}
	}

	newState := serializeFlexibleGPU(numspotFlexibleGpu)

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *flexibleGpuResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteFlexibleGPU(ctx, r.Provider, state.Id.ValueString(), vmIDFromModel(&state)); err != nil {
		response.Diagnostics.AddError("unable to delete flexible_gpu", err.Error())
		return
	}
}

func serializeFlexibleGPU(http *api.FlexibleGpu) resource_flexible_gpu.FlexibleGpuModel {
	return resource_flexible_gpu.FlexibleGpuModel{
		AvailabilityZoneName: types.StringValue(utils.ConvertAzNamePtrToString(http.AvailabilityZoneName)),
		DeleteOnVmDeletion:   types.BoolPointerValue(http.DeleteOnVmDeletion),
		Generation:           types.StringPointerValue(http.Generation),
		Id:                   types.StringPointerValue(http.Id),
		ModelName:            types.StringPointerValue(http.ModelName),
		State:                types.StringPointerValue(http.State),
		VmId:                 types.StringPointerValue(http.VmId),
	}
}

func deserializeCreateFlexibleGPURequest(tf *resource_flexible_gpu.FlexibleGpuModel) *api.CreateFlexibleGpuJSONRequestBody {
	return &api.CreateFlexibleGpuJSONRequestBody{
		AvailabilityZoneName: api.AvailabilityZoneName(tf.AvailabilityZoneName.ValueString()),
		DeleteOnVmDeletion:   tf.DeleteOnVmDeletion.ValueBoolPointer(),
		Generation:           utils.PointerOf(api.CreateFlexibleGpuGeneration(tf.Generation.ValueString())),
		ModelName:            api.CreateFlexibleGpuModelName(tf.ModelName.ValueString()),
	}
}
