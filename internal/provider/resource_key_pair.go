package provider

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_key_pair"
)

var (
	_ resource.Resource                = &KeyPairResource{}
	_ resource.ResourceWithConfigure   = &KeyPairResource{}
	_ resource.ResourceWithImportState = &KeyPairResource{}
)

type KeyPairResource struct {
	client *api.ClientWithResponses
}

func NewKeyPairResource() resource.Resource {
	return &KeyPairResource{}
}

func (r *KeyPairResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*api.ClientWithResponses)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *KeyPairResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *KeyPairResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_key_pair"
}

func (r *KeyPairResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_key_pair.KeyPairResourceSchema(ctx)
}

func (r *KeyPairResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_key_pair.KeyPairModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.CreateKeypairResponse, error) {
		body := KeyPairFromTfToCreateRequest(&data)
		return r.client.CreateKeypairWithResponse(ctx, body)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := KeyPairFromCreateHttpToTf(
		res.JSON200,
	)

	if !data.PublicKey.IsNull() && !data.PublicKey.IsUnknown() {
		tf.PublicKey = data.PublicKey
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *KeyPairResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_key_pair.KeyPairModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadKeypairsByIdResponse, error) {
		return r.client.ReadKeypairsByIdWithResponse(ctx, data.Id.ValueString()) // Use faker to inject token_200 status code
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := KeyPairFromReadHttpToTf(
		res.JSON200,
	)

	if !utils.IsTfValueNull(data.PublicKey) {
		tf.PublicKey = data.PublicKey
	}

	if !utils.IsTfValueNull(data.PrivateKey) {
		tf.PrivateKey = data.PrivateKey
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *KeyPairResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("implement me")
}

func (r *KeyPairResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_key_pair.KeyPairModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	utils.ExecuteRequest(func() (*api.DeleteKeypairResponse, error) {
		return r.client.DeleteKeypairWithResponse(ctx, data.Id.ValueString()) // Use faker to inject token_200 status code
	}, http.StatusOK, &response.Diagnostics)
}
