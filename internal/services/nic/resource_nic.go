package nic

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &NicResource{}
	_ resource.ResourceWithConfigure   = &NicResource{}
	_ resource.ResourceWithImportState = &NicResource{}
)

type NicResource struct {
	provider *client.NumSpotSDK
}

func NewNicResource() resource.Resource {
	return &NicResource{}
}

func (r *NicResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *NicResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *NicResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_nic"
}

func (r *NicResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = NicResourceSchema(ctx)
}

func (r *NicResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan NicModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		NicFromTfToCreateRequest(ctx, &plan, &response.Diagnostics),
		numspotClient.CreateNicWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create NIC", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(plan.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, numspotClient, r.provider.SpaceID, &response.Diagnostics, createdId, plan.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	tf := r.refreshNICState(ctx, createdId, []string{"attaching"}, []string{"available", "in-use"}, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	if !utils.IsTfValueNull(plan.LinkNic.VmId) && !utils.IsTfValueNull(plan.LinkNic.DeviceNumber) {
		tf = r.linkNIC(ctx, &plan, tf, &response.Diagnostics)

		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *NicResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.ReadNicsByIdResponse, error) {
		return numspotClient.ReadNicsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := NicFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NicResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan NicModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	nicId := state.Id.ValueString()

	// Update tags
	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			numspotClient,
			r.provider.SpaceID,
			nicId,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if !utils.IsTfValueNull(plan.LinkNic) &&
		!utils.IsTfValueNull(plan.LinkNic.VmId) && !state.LinkNic.VmId.Equal(plan.LinkNic.VmId) ||
		!utils.IsTfValueNull(plan.LinkNic.DeviceNumber) && !state.LinkNic.DeviceNumber.Equal(plan.LinkNic.DeviceNumber) {
		_ = r.updateLinkNIC(ctx, &plan, &state, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Update Nic
	if r.shouldUpdate(plan, state) {
		body := NicFromTfToUpdaterequest(ctx, &plan, &response.Diagnostics)
		res, err := numspotClient.UpdateNicWithResponse(ctx, r.provider.SpaceID, nicId, body)
		if err != nil {
			response.Diagnostics.AddError("failed to update nic", err.Error())
			return
		}
		if res.JSON200 == nil {
			response.Diagnostics.AddError("failed to update nic", utils.HandleError(res.Body).Error())
			return
		}
	}

	// Read resource
	res := utils.ExecuteRequest(func() (*numspot.ReadNicsByIdResponse, error) {
		return numspotClient.ReadNicsByIdWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := NicFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NicResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	if !utils.IsTfValueNull(data.LinkNic) {
		tf := r.unlinkNIC(ctx, &data, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
		response.Diagnostics.Append(request.State.Set(ctx, tf)...)
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), numspotClient.DeleteNicWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Nic", err.Error())
		return
	}
}

func (r *NicResource) updateLinkNIC(ctx context.Context, plan, state *NicModel, diags *diag.Diagnostics) *NicModel {
	var tf *NicModel
	if state.LinkNic.VmId.Equal(plan.LinkNic.VmId) &&
		state.LinkNic.DeviceNumber.Equal(plan.LinkNic.DeviceNumber) {
		return state
	}

	if !utils.IsTfValueNull(state.LinkNic) {
		tf = r.unlinkNIC(ctx, state, diags)
		if diags.HasError() {
			return nil
		}
	}

	if !utils.IsTfValueNull(plan.LinkNic) {
		tf = r.linkNIC(ctx, plan, tf, diags)
		if diags.HasError() {
			return nil
		}
	}

	return tf
}

func (r *NicResource) linkNIC(ctx context.Context, plan, state *NicModel, diags *diag.Diagnostics) *NicModel {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}

	resLink := utils.ExecuteRequest(func() (*numspot.LinkNicResponse, error) {
		return numspotClient.LinkNicWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString(), NicFromTfToLinkNICRequest(plan))
	}, http.StatusOK, diags)
	if resLink == nil {
		return nil
	}

	tf := r.refreshNICState(ctx, state.Id.ValueString(), []string{"available"}, []string{"in-use"}, diags)
	if diags.HasError() {
		return nil
	}

	return tf
}

func (r *NicResource) unlinkNIC(ctx context.Context, state *NicModel, diags *diag.Diagnostics) *NicModel {
	payload := numspot.UnlinkNicJSONRequestBody{LinkNicId: state.LinkNic.Id.ValueString()}
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}
	resLink := utils.ExecuteRequest(func() (*numspot.UnlinkNicResponse, error) {
		return numspotClient.UnlinkNicWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString(), payload)
	}, http.StatusNoContent, diags)
	if resLink == nil {
		return nil
	}

	tf := r.refreshNICState(ctx, state.Id.ValueString(), []string{"in-use"}, []string{"available"}, diags)
	if diags.HasError() {
		return nil
	}

	return tf
}

func (r *NicResource) refreshNICState(ctx context.Context, id string, startState, targetState []string, diags *diag.Diagnostics) *NicModel {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}

	// Retries read on resource until state is OK
	readRes, err := utils.RetryReadUntilStateValid(
		ctx,
		id,
		r.provider.SpaceID,
		startState,
		targetState,
		numspotClient.ReadNicsByIdWithResponse,
	)
	if err != nil {
		diags.AddError("Failed to read NIC", fmt.Sprintf("Error waiting for instance (%s) to be created/updated: %s", id, err))
		return nil
	}

	read, ok := readRes.(*numspot.Nic)
	if !ok {
		diags.AddError("Failed to read NIC", "object conversion error")
		return nil
	}

	tf := NicFromHttpToTf(ctx, read, diags)
	if diags.HasError() {
		return nil
	}

	return tf
}

func (r *NicResource) shouldUpdate(plan, state NicModel) bool {
	shouldUpdate := false
	shouldUpdate = shouldUpdate || (!utils.IsTfValueNull(plan.SecurityGroupIds) && !plan.SecurityGroupIds.Equal(state.SecurityGroupIds))
	shouldUpdate = shouldUpdate || (!utils.IsTfValueNull(plan.Description) && !plan.Description.Equal(state.Description))
	shouldUpdate = shouldUpdate || (!utils.IsTfValueNull(plan.LinkNic) && !plan.LinkNic.Equal(state.LinkNic))

	return shouldUpdate
}
