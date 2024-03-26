package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_subnet"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &SubnetResource{}
	_ resource.ResourceWithConfigure   = &SubnetResource{}
	_ resource.ResourceWithImportState = &SubnetResource{}
)

type SubnetResource struct {
	provider Provider
}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

func (r *SubnetResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *SubnetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *SubnetResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_subnet.SubnetResourceSchema(ctx)
}

func (r *SubnetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_subnet.SubnetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.CreateSubnetResponse, error) {
		body := SubnetFromTfToCreateRequest(&data)
		return r.provider.ApiClient.CreateSubnetWithResponse(ctx, r.provider.SpaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	createdId := *res.JSON201.Id
	createStateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"available"},
		Refresh: func() (result interface{}, state string, err error) {
			res, err := r.provider.ApiClient.ReadSubnetsByIdWithResponse(ctx, r.provider.SpaceID, createdId)
			if err != nil {
				response.Diagnostics.AddError("Failed to read Subnet", err.Error())
				return
			}

			if res.StatusCode() != http.StatusOK {
				apiError := utils.HandleError(res.Body)
				response.Diagnostics.AddError("Failed to read Subnet", apiError.Error())
				return
			}

			return res.JSON200, *res.JSON200.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   3 * time.Second,
	}

	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Net", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", *res.JSON201.Id, err))
		return
	}

	if data.MapPublicIpOnLaunch.ValueBool() {
		updateRes := utils.ExecuteRequest(func() (*iaas.UpdateSubnetResponse, error) {
			return r.provider.ApiClient.UpdateSubnetWithResponse(ctx, r.provider.SpaceID, createdId, iaas.UpdateSubnetJSONRequestBody{
				MapPublicIpOnLaunch: true,
			})
		}, http.StatusOK, &response.Diagnostics)
		if updateRes == nil {
			return
		}
	}

	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.ApiClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	readRes := utils.ExecuteRequest(func() (*iaas.ReadSubnetsByIdResponse, error) {
		return r.provider.ApiClient.ReadSubnetsByIdWithResponse(ctx, r.provider.SpaceID, createdId)
	}, http.StatusOK, &response.Diagnostics)
	if readRes == nil {
		return
	}

	tf, diags := SubnetFromHttpToTf(ctx, readRes.JSON200)
	if diags.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SubnetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_subnet.SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadSubnetsByIdResponse, error) {
		return r.provider.ApiClient.ReadSubnetsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := SubnetFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SubnetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resource_subnet.SubnetModel

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

	res := utils.ExecuteRequest(func() (*iaas.ReadSubnetsByIdResponse, error) {
		return r.provider.ApiClient.ReadSubnetsByIdWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := SubnetFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		return
	}
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SubnetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_subnet.SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.DeleteSubnetResponse, error) {
		return r.provider.ApiClient.DeleteSubnetWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}

	deleteStateConf := &retry.StateChangeConf{
		Pending: []string{"pending", "available"},
		Target:  []string{"deleted"},
		Refresh: func() (result interface{}, state string, err error) {
			res, err := r.provider.ApiClient.ReadSubnetsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
			if err != nil {
				response.Diagnostics.AddError("Failed to read Subnet on delete", err.Error())
				return
			}

			if res.StatusCode() != http.StatusOK {
				apiError := utils.HandleError(res.Body)
				if res.StatusCode() == http.StatusNotFound {
					return data, "deleted", nil
				}
				response.Diagnostics.AddError("Failed to read Subnet on delete", apiError.Error())
				return
			}

			return data, *res.JSON200.State, nil
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
