package nic

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &NicResource{}
	_ resource.ResourceWithConfigure   = &NicResource{}
	_ resource.ResourceWithImportState = &NicResource{}
)

type NicResource struct {
	provider services.IProvider
}

func NewNicResource() resource.Resource {
	return &NicResource{}
}

func (r *NicResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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
	var data NicModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils2.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		NicFromTfToCreateRequest(ctx, &data),
		r.provider.GetNumspotClient().CreateNicWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create NIC", err.Error())
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
	readRes, err := utils2.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.GetSpaceID(),
		[]string{"attaching"},
		[]string{"available", "in-use"},
		r.provider.GetNumspotClient().ReadNicsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Nic", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	read, ok := readRes.(*numspot.Nic)
	if !ok {
		response.Diagnostics.AddError("Failed to create nic", "object conversion error")
		return
	}

	tf, diagnostics := NicFromHttpToTf(ctx, read)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NicResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils2.ExecuteRequest(func() (*numspot.ReadNicsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadNicsByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := NicFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
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

	nicId := state.Id.ValueString()

	// Update tags
	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			nicId,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Update Nic
	updatedRes := utils2.ExecuteRequest(func() (*numspot.UpdateNicResponse, error) {
		body := NicFromTfToUpdaterequest(ctx, &plan, &response.Diagnostics)
		return r.provider.GetNumspotClient().UpdateNicWithResponse(ctx, r.provider.GetSpaceID(), nicId, body)
	}, http.StatusOK, &response.Diagnostics)

	if updatedRes == nil || response.Diagnostics.HasError() {
		return
	}

	// Read resource
	res := utils2.ExecuteRequest(func() (*numspot.ReadNicsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadNicsByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := NicFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NicResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils2.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteNicWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Nic", err.Error())
		return
	}
}
