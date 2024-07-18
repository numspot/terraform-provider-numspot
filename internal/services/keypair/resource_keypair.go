package keypair

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &KeyPairResource{}
	_ resource.ResourceWithConfigure   = &KeyPairResource{}
	_ resource.ResourceWithImportState = &KeyPairResource{}
)

type KeyPairResource struct {
	provider services.IProvider
}

func NewKeyPairResource() resource.Resource {
	return &KeyPairResource{}
}

func (r *KeyPairResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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
	var data KeyPairModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils2.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		KeyPairFromTfToCreateRequest(&data),
		r.provider.GetNumspotClient().CreateKeypairWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create KeyPair", err.Error())
		return
	}

	tf := KeyPairFromCreateHttpToTf(
		res.JSON201,
	)

	if !data.PublicKey.IsNull() && !data.PublicKey.IsUnknown() {
		tf.PublicKey = data.PublicKey
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *KeyPairResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data KeyPairModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils2.ExecuteRequest(func() (*numspot.ReadKeypairsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadKeypairsByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString()) // Use faker to inject token_200 status code
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := KeyPairFromReadHttpToTf(
		res.JSON200,
	)

	if !utils2.IsTfValueNull(data.PublicKey) {
		tf.PublicKey = data.PublicKey
	}

	if !utils2.IsTfValueNull(data.PrivateKey) {
		tf.PrivateKey = data.PrivateKey
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *KeyPairResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("implement me")
}

func (r *KeyPairResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data KeyPairModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils2.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteKeypairWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete KeyPair", err.Error())
		return
	}
}
