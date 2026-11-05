package keypair

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/keypair/resource_keypair"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &keypairResource{}
	_ resource.ResourceWithConfigure   = &keypairResource{}
	_ resource.ResourceWithImportState = &keypairResource{}
)

// keypairResource is the Terraform resource implementation for numspot_keypair.
type keypairResource struct {
	services.BaseResource
}

// NewKeypairResource returns a new keypairResource instance.
func NewKeypairResource() resource.Resource {
	return &keypairResource{}
}

// Metadata sets the resource type name.
func (r *keypairResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_keypair"
}

// Schema defines the schema for the resource.
func (r *keypairResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_keypair.KeypairResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *keypairResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_keypair.KeypairModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numspotKeypair, err := core.CreateKeypair(ctx, r.Provider, deserializeCreateKeypairRequest(&plan))
	if err != nil {
		response.Diagnostics.AddError("unable to create keypair", err.Error())
		return
	}

	state := serializeCreateKeypair(numspotKeypair)
	state.PublicKey = utils.PreserveFromState(plan.PublicKey)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *keypairResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_keypair.KeypairModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotKeypair, err := core.ReadKeypair(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read keypair", err.Error())
		return
	}

	newState := serializeKeypair(numspotKeypair)
	newState.PrivateKey = state.PrivateKey
	newState.PublicKey = state.PublicKey

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *keypairResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *keypairResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_keypair.KeypairModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteKeypair(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete keypair", err.Error())
		return
	}
}

func serializeKeypair(http *api.Keypair) resource_keypair.KeypairModel {
	return resource_keypair.KeypairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		Type:        types.StringPointerValue(http.Type),
	}
}

func deserializeCreateKeypairRequest(tf *resource_keypair.KeypairModel) *api.CreateKeypairJSONRequestBody {
	return &api.CreateKeypairJSONRequestBody{
		Name:      tf.Name.ValueString(),
		PublicKey: tf.PublicKey.ValueStringPointer(),
	}
}

func serializeCreateKeypair(http *api.CreateKeypair) resource_keypair.KeypairModel {
	return resource_keypair.KeypairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
		Type:        types.StringPointerValue(http.Type),
	}
}
