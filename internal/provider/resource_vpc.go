package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpc"
)

var (
	_ resource.Resource                = &VpcResource{}
	_ resource.ResourceWithConfigure   = &VpcResource{}
	_ resource.ResourceWithImportState = &VpcResource{}
)

type VpcResource struct {
	provider Provider
}

func NewNetResource() resource.Resource {
	return &VpcResource{}
}

func (r *VpcResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *VpcResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VpcResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpc"
}

func (r *VpcResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_vpc.VpcResourceSchema(ctx)
}

func (r *VpcResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_vpc.VpcModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.CreateVpcResponse, error) {
		body := NetFromTfToCreateRequest(&data)
		return r.provider.ApiClient.CreateVpcWithResponse(ctx, r.provider.SpaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	// Handle tags
	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.ApiClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	createStateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"available"},
		Refresh: func() (result interface{}, state string, err error) {
			readRes := utils.ExecuteRequest(func() (*api.ReadVpcsByIdResponse, error) {
				return r.provider.ApiClient.ReadVpcsByIdWithResponse(ctx, r.provider.SpaceID, createdId)
			}, http.StatusOK, &response.Diagnostics)
			if readRes == nil {
				return
			}

			return readRes.JSON200, *readRes.JSON200.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   3 * time.Second,
	}

	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Net", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", createdId, err))
		return
	}

	tf := NetFromHttpToTf(ctx, res.JSON201)
	tf.Tags = tags.ReadTags(ctx, r.provider.ApiClient, r.provider.SpaceID, response.Diagnostics, createdId)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VpcResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_vpc.VpcModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadVpcsByIdResponse, error) {
		return r.provider.ApiClient.ReadVpcsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := NetFromHttpToTf(ctx, res.JSON200)
	tf.Tags = tags.ReadTags(ctx, r.provider.ApiClient, r.provider.SpaceID, response.Diagnostics, data.Id.ValueString())
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VpcResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state resource_vpc.VpcModel
		plan  resource_vpc.VpcModel
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

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
	}

	res := utils.ExecuteRequest(func() (*api.ReadVpcsByIdResponse, error) {
		return r.provider.ApiClient.ReadVpcsByIdWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := NetFromHttpToTf(ctx, res.JSON200)
	tf.Tags = tags.ReadTags(ctx, r.provider.ApiClient, r.provider.SpaceID, response.Diagnostics, state.Id.ValueString())
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VpcResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_vpc.VpcModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.DeleteVpcResponse, error) {
		return r.provider.ApiClient.DeleteVpcWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}

	deleteStateConf := &retry.StateChangeConf{
		Pending: []string{"pending", "available", "deleting"},
		Target:  []string{"deleted"},
		Refresh: func() (result interface{}, state string, err error) {
			// Do not use utils.ExecuteRequest to access error response to know if it's a 404 Not Found expected response
			readNetRes, err := r.provider.ApiClient.ReadVpcsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
			if err != nil {
				response.Diagnostics.AddError("Failed to read Net on delete", err.Error())
				return
			}

			if readNetRes.StatusCode() != http.StatusOK {
				apiError := utils.HandleError(readNetRes.Body)
				if readNetRes.StatusCode() == http.StatusNotFound {
					return data, "deleted", nil
				}
				response.Diagnostics.AddError("Failed to read Net on delete", apiError.Error())
				return
			}

			return data, *readNetRes.JSON200.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   5 * time.Second,
	}

	_, err := deleteStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Net", fmt.Sprintf("Error waiting for instance (%s) to be deleted: %s", data.Id.ValueString(), err))
		return
	}

	response.State.RemoveResource(ctx)
}
