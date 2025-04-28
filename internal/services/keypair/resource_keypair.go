package keypair

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/keypair/resource_keypair"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &keypairResource{}
	_ resource.ResourceWithConfigure   = &keypairResource{}
	_ resource.ResourceWithImportState = &keypairResource{}
)

type keypairResource struct {
	provider *client.NumSpotSDK
}

func NewKeyPairResource() resource.Resource {
	return &keypairResource{}
}

func (r *keypairResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *keypairResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *keypairResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_keypair"
}

func (r *keypairResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_keypair.KeypairResourceSchema(ctx)
}

func (r *keypairResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_keypair.KeypairModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	keypair, err := core.CreateKeypair(ctx, r.provider, deserializeCreateNumSpotKeypair(plan))
	if err != nil {
		response.Diagnostics.AddError("unable to create keypair", err.Error())
		return
	}

	state := serializeNumSpotCreateKeypair(keypair)
	if !utils.IsTfValueNull(plan.PublicKey) {
		state.PublicKey = plan.PublicKey
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *keypairResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_keypair.KeypairModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	keypairID := state.Id.ValueString()

	numSpotKeypair, err := core.ReadKeypair(ctx, r.provider, keypairID)
	if err != nil {
		response.Diagnostics.AddError("unable to read keypair", err.Error())
		return
	}

	newState := serializeNumSpotKeypair(numSpotKeypair)

	if state.PublicKey.ValueString() != "" {
		newState.PublicKey = state.PublicKey
	}
	if state.PrivateKey.ValueString() != "" {
		newState.PrivateKey = state.PrivateKey
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *keypairResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *keypairResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_keypair.KeypairModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	if err := core.DeleteKeypair(ctx, r.provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete keypair", err.Error())
		return
	}
}

func serializeNumSpotCreateKeypair(http *api.CreateKeypair) resource_keypair.KeypairModel {
	return resource_keypair.KeypairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
	}
}

func serializeNumSpotKeypair(http *api.Keypair) resource_keypair.KeypairModel {
	return resource_keypair.KeypairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
	}
}

func deserializeCreateNumSpotKeypair(tf resource_keypair.KeypairModel) api.CreateKeypairJSONRequestBody {
	return api.CreateKeypairJSONRequestBody{
		Name:      tf.Name.ValueString(),
		PublicKey: tf.PublicKey.ValueStringPointer(),
	}
}
