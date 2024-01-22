package provider

import (
	"context"
	"fmt"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_public_ip"
)

var (
	_ resource.Resource                = &PublicIpResource{}
	_ resource.ResourceWithConfigure   = &PublicIpResource{}
	_ resource.ResourceWithImportState = &PublicIpResource{}
)

type PublicIpResource struct {
	client *api.ClientWithResponses
}

func NewPublicIpResource() resource.Resource {
	return &PublicIpResource{}
}

func (r *PublicIpResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *PublicIpResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	var data resource_public_ip.PublicIpModel
	response.State.Get(ctx, &data)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *PublicIpResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_public_ip"
}

func (r *PublicIpResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_public_ip.PublicIpResourceSchema(ctx)
}

func (r *PublicIpResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	body := PublicIpFromTfToCreateRequest(data)
	res, err := r.client.CreatePublicIpWithResponse(ctx, body)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to create PublicIp", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 201)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to create PublicIp", apiError.Error())
		return
	}

	tf := PublicIpFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)

	//Refresh from state
	//response.Diagnostics.Append(response.State.Get(ctx, &data)...)

	if tf.VmId.IsNull() && tf.NicId.IsNull() {
		return
	}

	//Call Link publicIP
	if err := invokeLinkPublicIP(ctx, r.client, tf); err != nil {
		response.Diagnostics.AddError("Failed to link public IP", err.Error())
	}
}

func (r *PublicIpResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement READ operation
	res, err := r.client.ReadPublicIpsByIdWithResponse(ctx, data.PublicIp.ValueString())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to read PublicIp", apiError.Error())
		return
	}

	tf := PublicIpFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *PublicIpResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.VmId.IsNull() && data.NicId.IsNull() {
		return
	}

	//Call Link publicIP
	if err := invokeLinkPublicIP(ctx, r.client, data); err != nil {
		response.Diagnostics.AddError("Failed to link public IP", err.Error())
	}
}

func (r *PublicIpResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// TODO: Implement DELETE operation
	res, err := r.client.DeletePublicIpWithResponse(ctx, data.PublicIp.ValueString(), api.DeletePublicIpRequestSchema{})
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete PublicIp", err.Error())
		return
	}

	expectedStatusCode := 200 // FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to delete PublicIp", apiError.Error())
		return
	}
}
