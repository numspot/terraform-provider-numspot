package directlink

/*

 DIRECT LINKS are not handled for now

import (
	"context"
	"fmt"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &DirectLinkResource{}
	_ resource.ResourceWithConfigure   = &DirectLinkResource{}
	_ resource.ResourceWithImportState = &DirectLinkResource{}
)

type DirectLinkResource struct {
	provider services.IProvider
}

func NewDirectLinkResource() resource.Resource {
	return &DirectLinkResource{}
}

func (r *DirectLinkResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *DirectLinkResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *DirectLinkResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_direct_link"
}

func (r *DirectLinkResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = DirectLinkResourceSchema(ctx)
}

func (r *DirectLinkResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data DirectLinkModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		DirectLinkFromTfToCreateRequest(&data),
		r.provider.GetNumspotClient().CreateDirectLinkWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Direct Link", err.Error())
		return
	}

	// Retries read on resource until state is OK
	createdId := *res.JSON201.Id
	_, err = utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.GetSpaceID(),
		[]string{"pending", "requested"},
		[]string{"available"},
		r.provider.GetNumspotClient().ReadDirectLinksByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Direct Link", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	tf := DirectLinkFromHttpToTf(res.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data DirectLinkModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*numspot.ReadDirectLinksByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadDirectLinksByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.String())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := DirectLinkFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DirectLinkResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("implement me")
}

func (r *DirectLinkResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data DirectLinkModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteDirectLinkWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Direct Link", err.Error())
		return
	}
}
*/
