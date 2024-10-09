package natgateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &NatGatewayResource{}
	_ resource.ResourceWithConfigure   = &NatGatewayResource{}
	_ resource.ResourceWithImportState = &NatGatewayResource{}
)

type NatGatewayResource struct {
	provider services.IProvider
}

func NewNatGatewayResource() resource.Resource {
	return &NatGatewayResource{}
}

func (r *NatGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *NatGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *NatGatewayResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_nat_gateway"
}

func (r *NatGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = NatGatewayResourceSchema(ctx)
}

func (r *NatGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data NatGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		NatGatewayFromTfToCreateRequest(data),
		r.provider.GetNumspotClient().CreateNatGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Nat Gateway", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Retries read on resource until state is OK
	read, err := utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.GetSpaceID(),
		[]string{"pending"},
		[]string{"available"},
		r.provider.GetNumspotClient().ReadNatGatewayByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Nat Gateway", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	rr, ok := read.(*numspot.NatGateway)
	if !ok {
		response.Diagnostics.AddError("Failed to create nat gateway", "object conversion error")
		return
	}

	tf := NatGatewayFromHttpToTf(ctx, rr, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NatGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data NatGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*numspot.ReadNatGatewayByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadNatGatewayByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := NatGatewayFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NatGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan NatGatewayModel
	modifications := false
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			state.Id.ValueString(),
		)

		if response.Diagnostics.HasError() {
			return
		}
		modifications = true
	}

	if !modifications {
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.ReadNatGatewayByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadNatGatewayByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := NatGatewayFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NatGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data NatGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteNatGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete NAT Gateway", err.Error())
		return
	}
}
