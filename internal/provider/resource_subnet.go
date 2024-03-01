package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_subnet"
)

var (
	_ resource.Resource                = &SubnetResource{}
	_ resource.ResourceWithConfigure   = &SubnetResource{}
	_ resource.ResourceWithImportState = &SubnetResource{}
)

type SubnetResource struct {
	client *api.ClientWithResponses
}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

func (r *SubnetResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*api.ClientWithResponses)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.client = client
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

	res := utils.ExecuteRequest(func() (*api.CreateSubnetResponse, error) {
		body := SubnetFromTfToCreateRequest(&data)
		return r.client.CreateSubnetWithResponse(ctx, spaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	createdId := *res.JSON201.Id
	createStateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"available"},
		Refresh: func() (result interface{}, state string, err error) {
			res, err := r.client.ReadSubnetsByIdWithResponse(ctx, spaceID, createdId)
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
		updateRes := utils.ExecuteRequest(func() (*api.UpdateSubnetResponse, error) {
			return r.client.UpdateSubnetWithResponse(ctx, spaceID, createdId, api.UpdateSubnetJSONRequestBody{
				MapPublicIpOnLaunch: true,
			})
		}, http.StatusOK, &response.Diagnostics)
		if updateRes == nil {
			return
		}
	}

	readRes := utils.ExecuteRequest(func() (*api.ReadSubnetsByIdResponse, error) {
		return r.client.ReadSubnetsByIdWithResponse(ctx, spaceID, createdId)
	}, http.StatusOK, &response.Diagnostics)
	if readRes == nil {
		return
	}

	tf := SubnetFromHttpToTf(readRes.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SubnetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_subnet.SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadSubnetsByIdResponse, error) {
		return r.client.ReadSubnetsByIdWithResponse(ctx, spaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := SubnetFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SubnetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("implement me")
}

func (r *SubnetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_subnet.SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.DeleteSubnetResponse, error) {
		return r.client.DeleteSubnetWithResponse(ctx, spaceID, data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}

	deleteStateConf := &retry.StateChangeConf{
		Pending: []string{"pending", "available"},
		Target:  []string{"deleted"},
		Refresh: func() (result interface{}, state string, err error) {
			res, err := r.client.ReadSubnetsByIdWithResponse(ctx, spaceID, data.Id.ValueString())
			if err != nil {
				response.Diagnostics.AddError("Failed to read Subnet on delete", err.Error())
				return
			}

			expectedStatusCode := 200
			if res.StatusCode() != expectedStatusCode {
				apiError := utils.HandleError(res.Body)
				if match, _ := regexp.MatchString("No .* found", apiError.Error()); match {
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
