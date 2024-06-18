package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_virtual_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VirtualGatewayResource{}
	_ resource.ResourceWithConfigure   = &VirtualGatewayResource{}
	_ resource.ResourceWithImportState = &VirtualGatewayResource{}
)

type VirtualGatewayResource struct {
	provider Provider
}

func NewVirtualGatewayResource() resource.Resource {
	return &VirtualGatewayResource{}
}

func (r *VirtualGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(Provider)
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
	response.Schema = resource_virtual_gateway.VirtualGatewayResourceSchema(ctx)
}

func (r *VirtualGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		VirtualGatewayFromTfToCreateRequest(data),
		r.provider.ApiClient.CreateVirtualGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Virtual Gateway", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.ApiClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Link virtual gateway to VPCs
	if !data.VpcId.IsNull() {
		_ = utils.ExecuteRequest(func() (*iaas.LinkVirtualGatewayToVpcResponse, error) {
			return r.provider.ApiClient.LinkVirtualGatewayToVpcWithResponse(
				ctx,
				r.provider.SpaceID,
				createdId,
				iaas.LinkVirtualGatewayToVpcJSONRequestBody{
					VpcId: data.VpcId.ValueString(),
				},
			)
		}, http.StatusOK, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Retries read on resource until state is OK
	read, err := retry_utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"available"},
		r.provider.ApiClient.ReadVirtualGatewaysByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Virtual Gateway", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	rr, ok := read.(*iaas.VirtualGateway)
	if !ok {
		response.Diagnostics.AddError("Failed to create virtual gateway", "object conversion error")
		return
	}

	tf, diagnostics := VirtualGatewayFromHttpToTf(ctx, rr)
	if diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VirtualGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadVirtualGatewaysByIdResponse, error) {
		return r.provider.ApiClient.ReadVirtualGatewaysByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := VirtualGatewayFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VirtualGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan resource_virtual_gateway.VirtualGatewayModel
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
			r.provider.ApiClient,
			r.provider.SpaceID,
			state.Id.ValueString(),
		)
		if response.Diagnostics.HasError() {
			return
		}
		modifications = true
	}

	// Link/Unlink virtual gateway to VPCs
	if state.VpcId.ValueString() != plan.VpcId.ValueString() {

		// Unlink
		if !(state.VpcId.IsNull() || state.VpcId.IsUnknown()) {
			_ = utils.ExecuteRequest(func() (*iaas.UnlinkVirtualGatewayToVpcResponse, error) {
				return r.provider.ApiClient.UnlinkVirtualGatewayToVpcWithResponse(
					ctx,
					r.provider.SpaceID,
					state.Id.ValueString(),
					iaas.UnlinkVirtualGatewayToVpcJSONRequestBody{
						VpcId: state.VpcId.ValueString(),
					},
				)
			}, http.StatusNoContent, &response.Diagnostics)
			if response.Diagnostics.HasError() {
				return
			}
		}

		// Link
		if !(plan.VpcId.IsNull() || plan.VpcId.IsUnknown()) {
			_ = utils.ExecuteRequest(func() (*iaas.LinkVirtualGatewayToVpcResponse, error) {
				return r.provider.ApiClient.LinkVirtualGatewayToVpcWithResponse(
					ctx,
					r.provider.SpaceID,
					state.Id.ValueString(),
					iaas.LinkVirtualGatewayToVpcJSONRequestBody{
						VpcId: plan.VpcId.ValueString(),
					},
				)
			}, http.StatusOK, &response.Diagnostics)
			if response.Diagnostics.HasError() {
				return
			}
		}

		modifications = true
	}

	if !modifications {
		return
	}

	res := utils.ExecuteRequest(func() (*iaas.ReadVirtualGatewaysByIdResponse, error) {
		return r.provider.ApiClient.ReadVirtualGatewaysByIdWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := VirtualGatewayFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VirtualGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// Unlink
	if !(data.VpcId.IsNull() || data.VpcId.IsUnknown()) {
		_ = utils.ExecuteRequest(func() (*iaas.UnlinkVirtualGatewayToVpcResponse, error) {
			return r.provider.ApiClient.UnlinkVirtualGatewayToVpcWithResponse(
				ctx,
				r.provider.SpaceID,
				data.Id.ValueString(),
				iaas.UnlinkVirtualGatewayToVpcJSONRequestBody{
					VpcId: data.VpcId.ValueString(),
				},
			)
		}, http.StatusNoContent, &response.Diagnostics)

		// Note : don't return in case of error, we want to try to delete the resource anyway
	}

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteVirtualGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Virtual Gateway", err.Error())
		return
	}
}
