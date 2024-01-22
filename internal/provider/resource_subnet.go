package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
	"net/http"
	"time"

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

	body := SubnetFromTfToCreateRequest(data)
	res, err := r.client.CreateSubnetWithResponse(ctx, body)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to create Subnet", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 201)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to create Subnet", apiError.Error())
		return
	}

	createdId := *res.JSON200.Id
	createStateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"available"},
		Refresh: func() (result interface{}, state string, err error) {
			res, err := r.client.ReadSubnetsByIdWithResponse(ctx, createdId)
			if err != nil {
				response.Diagnostics.AddError("Failed to read Subnet", err.Error())
			}

			expectedStatusCode := 200
			if res.StatusCode() != expectedStatusCode {
				// TODO: Handle NumSpot error
				apiError := utils.HandleError(res.Body)
				response.Diagnostics.AddError("Failed to read Subnet", apiError.Error())
				return
			}

			return res.JSON200, *res.JSON200.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   3 * time.Second,
	}

	_, err = createStateConf.WaitForStateContext(ctx)
	if err != nil {
		// return fmt.Errorf("Error waiting for example instance (%s) to be created: %s", d.Id(), err)
		response.Diagnostics.AddError("Failed to create Net", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", *res.JSON200.Id, err))
		return
	}

	if data.MapPublicIpOnLaunch.ValueBool() {
		updateRes, err := r.client.UpdateSubnetWithResponse(ctx, createdId, api.UpdateSubnetJSONRequestBody{
			MapPublicIpOnLaunch: true,
		})

		if err != nil {
			response.Diagnostics.AddError("Failed to read Subnet", err.Error())
			return
		}

		if updateRes.StatusCode() != http.StatusOK {
			apiError := utils.HandleError(res.Body)
			response.Diagnostics.AddError("Failed to read Subnet", apiError.Error())
			return
		}
	}

	readRes, err := r.client.ReadSubnetsByIdWithResponse(ctx, createdId)
	if err != nil {
		response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
		return
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(readRes.Body)
		response.Diagnostics.AddError("Failed to read Subnet", apiError.Error())
		return
	}

	tf := SubnetFromHttpToTf(readRes.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SubnetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_subnet.SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement READ operation
	res, err := r.client.ReadSubnetsByIdWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to read Subnet", apiError.Error())
		return
	}

	tf := SubnetFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *SubnetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *SubnetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_subnet.SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// TODO: Implement DELETE operation
	res, err := r.client.DeleteSubnetWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete Subnet", err.Error())
		return
	}

	expectedStatusCode := 200 // FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to delete Subnet", apiError.Error())
		return
	}

	deleteStateConf := &retry.StateChangeConf{
		Pending: []string{"pending", "available"},
		Target:  []string{"deleted"},
		Refresh: func() (result interface{}, state string, err error) {
			res, err := r.client.ReadSubnetsByIdWithResponse(ctx, data.Id.ValueString())
			if err != nil {
				response.Diagnostics.AddError("Failed to read Subnet on delete", err.Error())
			}

			expectedStatusCode := 200
			if res.StatusCode() != expectedStatusCode {
				// TODO: Handle NumSpot error
				apiError := utils.HandleError(res.Body)
				if apiError.Error() == "No Subnets found" {
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

	_, err = deleteStateConf.WaitForStateContext(ctx)
	if err != nil {
		// return fmt.Errorf("Error waiting for example instance (%s) to be created: %s", d.Id(), err)
		response.Diagnostics.AddError("Failed to delete Net", fmt.Sprintf("Error waiting for instance (%s) to be deleted: %s", data.Id.ValueString(), err))
		return
	}

	response.State.RemoveResource(ctx)
}
