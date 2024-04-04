package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_public_ip"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &PublicIpResource{}
	_ resource.ResourceWithConfigure   = &PublicIpResource{}
	_ resource.ResourceWithImportState = &PublicIpResource{}
)

type PublicIpResource struct {
	provider Provider
}

func NewPublicIpResource() resource.Resource {
	return &PublicIpResource{}
}

func (r *PublicIpResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(Provider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *PublicIpResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *PublicIpResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_public_ip"
}

func (r *PublicIpResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_public_ip.PublicIpResourceSchema(ctx)
}

func (r *PublicIpResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan, state resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	createRes := utils.ExecuteRequest(func() (*iaas.CreatePublicIpResponse, error) {
		return r.provider.ApiClient.CreatePublicIpWithResponse(ctx, r.provider.SpaceID)
	}, http.StatusCreated, &response.Diagnostics)
	if createRes == nil {
		return
	}

	tf := PublicIpFromHttpToTf(createRes.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)

	if plan.VmId.IsNull() && plan.NicId.IsUnknown() {
		return
	}

	state.VmId = plan.VmId
	state.NicId = plan.NicId

	// Call Link publicIP
	linkPublicIP, err := invokeLinkPublicIP(ctx, r.provider, &state)
	if err != nil {
		response.Diagnostics.AddError("Failed to link public IP", err.Error())
	}
	state.LinkPublicIP = types.StringPointerValue(linkPublicIP)

	// Refresh state
	data, err := refreshState(ctx, r.provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Failed to read PublicIp", err.Error())
	}
	response.Diagnostics.Append(response.State.Set(ctx, *data)...)
}

func (r *PublicIpResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	readRes := utils.ExecuteRequest(func() (*iaas.ReadPublicIpsByIdResponse, error) {
		return r.provider.ApiClient.ReadPublicIpsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if readRes == nil {
		return
	}

	tf := PublicIpFromHttpToTf(readRes.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *PublicIpResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		plan, state  resource_public_ip.PublicIpModel
		linkPublicIP *string
		err          error
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	data, err := refreshState(ctx, r.provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Failed to read PublicIp", err.Error())
		return
	}

	chgSet := ComputePublicIPChangeSet(&plan, data)

	if chgSet.Err != nil {
		response.Diagnostics.AddError("Failed to update public IP", err.Error()) // err ??
		return
	}

	if chgSet.Unlink {
		if err := invokeUnlinkPublicIP(ctx, r.provider, &state); err != nil {
			response.Diagnostics.AddError("Failed to unlink public IP", err.Error())
			return
		}
		state.LinkPublicIP = types.StringNull()
		data, err := refreshState(ctx, r.provider, state.Id.ValueString())
		if err != nil {
			response.Diagnostics.AddError("Failed to read PublicIp", err.Error())
			return
		}
		response.Diagnostics.Append(response.State.Set(ctx, *data)...)
	}
	if chgSet.Link {
		plan.Id = state.Id
		linkPublicIP, err = invokeLinkPublicIP(ctx, r.provider, &plan)
		if err != nil {
			response.Diagnostics.AddError("Failed to link public IP", err.Error())
			return
		}
		state.LinkPublicIP = types.StringPointerValue(linkPublicIP)
	}
	data, err = refreshState(ctx, r.provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Failed to read PublicIp", err.Error())
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, *data)...)
}

func (r *PublicIpResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if !state.LinkPublicIP.IsNull() {
		if err := invokeUnlinkPublicIP(ctx, r.provider, &state); err != nil {
			response.Diagnostics.AddError("Failed to unlink public IP", err.Error())
			return
		}
	}
	utils.ExecuteRequest(func() (*iaas.DeletePublicIpResponse, error) {
		return r.provider.ApiClient.DeletePublicIpWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
}
