package virtualgateway

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
	_ resource.Resource                = &VirtualGatewayResource{}
	_ resource.ResourceWithConfigure   = &VirtualGatewayResource{}
	_ resource.ResourceWithImportState = &VirtualGatewayResource{}
)

type VirtualGatewayResource struct {
	provider services.IProvider
}

func NewVirtualGatewayResource() resource.Resource {
	return &VirtualGatewayResource{}
}

func (r *VirtualGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(services.IProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = client
}

func (r *VirtualGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VirtualGatewayResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_virtual_gateway"
}

func (r *VirtualGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VirtualGatewayResourceSchema(ctx)
}

func (r *VirtualGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data VirtualGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		VirtualGatewayFromTfToCreateRequest(data),
		r.provider.GetNumspotClient().CreateVirtualGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Virtual Gateway", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Link virtual gateway to VPCs
	if !data.VpcId.IsNull() {
		_ = utils.ExecuteRequest(func() (*numspot.LinkVirtualGatewayToVpcResponse, error) {
			return r.provider.GetNumspotClient().LinkVirtualGatewayToVpcWithResponse(
				ctx,
				r.provider.GetSpaceID(),
				createdId,
				numspot.LinkVirtualGatewayToVpcJSONRequestBody{
					VpcId: data.VpcId.ValueString(),
				},
			)
		}, http.StatusOK, &response.Diagnostics)
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
		r.provider.GetNumspotClient().ReadVirtualGatewaysByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Virtual Gateway", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	rr, ok := read.(*numspot.VirtualGateway)
	if !ok {
		response.Diagnostics.AddError("Failed to create virtual gateway", "object conversion error")
		return
	}

	tf := VirtualGatewayFromHttpToTf(ctx, rr, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VirtualGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*numspot.ReadVirtualGatewaysByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadVirtualGatewaysByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := VirtualGatewayFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VirtualGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan VirtualGatewayModel
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
	res := utils.ExecuteRequest(func() (*numspot.ReadVirtualGatewaysByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadVirtualGatewaysByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := VirtualGatewayFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VirtualGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// Unlink
	if !(data.VpcId.IsNull() || data.VpcId.IsUnknown()) {
		_ = utils.ExecuteRequest(func() (*numspot.UnlinkVirtualGatewayToVpcResponse, error) {
			return r.provider.GetNumspotClient().UnlinkVirtualGatewayToVpcWithResponse(
				ctx,
				r.provider.GetSpaceID(),
				data.Id.ValueString(),
				numspot.UnlinkVirtualGatewayToVpcJSONRequestBody{
					VpcId: data.VpcId.ValueString(),
				},
			)
		}, http.StatusNoContent, &response.Diagnostics)

		// Note : don't return in case of error, we want to try to delete the resource anyway
	}

	err := utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteVirtualGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Virtual Gateway", err.Error())
		return
	}
}
