package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpn_connection"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VpnConnectionResource{}
	_ resource.ResourceWithConfigure   = &VpnConnectionResource{}
	_ resource.ResourceWithImportState = &VpnConnectionResource{}
)

type VpnConnectionResource struct {
	provider Provider
}

func NewVpnConnectionResource() resource.Resource {
	return &VpnConnectionResource{}
}

func (r *VpnConnectionResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *VpnConnectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VpnConnectionResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpn_connection"
}

func (r *VpnConnectionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_vpn_connection.VpnConnectionResourceSchema(ctx)
}

func (r *VpnConnectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res, err := utils.RetryCreateUntilResourceAvailable(
		ctx,
		r.provider.SpaceID,
		VpnConnectionFromTfToCreateRequest(&data),
		r.provider.ApiClient.CreateVpnConnectionWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPN Connection", err.Error())
		return
	}

	tf := VpnConnectionFromHttpToTf(ctx, res.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VpnConnectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadVpnConnectionsByIdResponse, error) {
		return r.provider.ApiClient.ReadVpnConnectionsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := VpnConnectionFromHttpToTf(ctx, res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VpnConnectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("nothing to do")
}

func (r *VpnConnectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteVpnConnectionWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete VPN Connection", err.Error())
		return
	}
}
