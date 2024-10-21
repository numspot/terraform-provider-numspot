package keypair

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &KeyPairResource{}
	_ resource.ResourceWithConfigure   = &KeyPairResource{}
	_ resource.ResourceWithImportState = &KeyPairResource{}
)

type KeyPairResource struct {
	provider *client.NumSpotSDK
}

func NewKeyPairResource() resource.Resource {
	return &KeyPairResource{}
}

func (r *KeyPairResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *KeyPairResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *KeyPairResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_keypair"
}

func (r *KeyPairResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = KeyPairResourceSchema(ctx)
}

func (r *KeyPairResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan KeyPairModel
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

func (r *KeyPairResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state KeyPairModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	keypairID := state.Id.ValueString()

	numSpotKeypair, err := core.ReadKeypairsWithID(ctx, r.provider, keypairID)
	if err != nil {
		response.Diagnostics.AddError("unable to read keypair", err.Error())
		return
	}

	newState := serializeNumSpotKeypair(numSpotKeypair)

	if !utils.IsTfValueNull(state.PublicKey) {
		newState.PublicKey = state.PublicKey
	}
	if !utils.IsTfValueNull(state.PrivateKey) {
		newState.PrivateKey = state.PrivateKey
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *KeyPairResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("no update for keypairs")
}

func (r *KeyPairResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state KeyPairModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	if err := core.DeleteKeypair(ctx, r.provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("failed to delete keypair", err.Error())
		return
	}
}

func serializeNumSpotCreateKeypair(http *numspot.CreateKeypair) KeyPairModel {
	return KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
		PrivateKey:  types.StringPointerValue(http.PrivateKey),
	}
}

func serializeNumSpotKeypair(http *numspot.Keypair) KeyPairModel {
	return KeyPairModel{
		Fingerprint: types.StringPointerValue(http.Fingerprint),
		Id:          types.StringPointerValue(http.Name),
		Name:        types.StringPointerValue(http.Name),
	}
}

func deserializeCreateNumSpotKeypair(tf KeyPairModel) numspot.CreateKeypairJSONRequestBody {
	return numspot.CreateKeypairJSONRequestBody{
		Name:      tf.Name.ValueString(),
		PublicKey: tf.PublicKey.ValueStringPointer(),
	}
}
