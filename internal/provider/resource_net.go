package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_net"
)

var (
	_ resource.Resource                = &NetResource{}
	_ resource.ResourceWithConfigure   = &NetResource{}
	_ resource.ResourceWithImportState = &NetResource{}
)

type NetResource struct {
	client *api.ClientWithResponses
}

func NewNetResource() resource.Resource {
	return &NetResource{}
}

func (r *NetResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *NetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *NetResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_net"
}

func (r *NetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_net.NetResourceSchema(ctx)
}

func (r *NetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_net.NetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	body := NetFromTfToCreateRequest(data)
	res, err := r.client.CreateNetWithResponse(ctx, body)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to create Net", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 201)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to create Net", apiError.Error())
		return
	}

	createdId := *res.JSON200.Id
	createStateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"available"},
		Refresh: func() (result interface{}, state string, err error) {
			res, err := r.client.ReadNetsByIdWithResponse(ctx, createdId)
			if err != nil {
				response.Diagnostics.AddError("Failed to read Net", err.Error())
			}

			expectedStatusCode := 200
			if res.StatusCode() != expectedStatusCode {
				// TODO: Handle NumSpot error
				apiError := utils.HandleError(res.Body)
				response.Diagnostics.AddError("Failed to read Net", apiError.Error())
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

	tf := NetFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_net.NetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement READ operation
	res, err := r.client.ReadNetsByIdWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to read Net", apiError.Error())
		return
	}

	tf := NetFromHttpToTf(res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *NetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_net.NetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// TODO: Implement DELETE operation
	res, err := r.client.DeleteNetWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete Net", err.Error())
		return
	}

	expectedStatusCode := 200 // FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to delete Net", apiError.Error())
		return
	}

	deleteStateConf := &retry.StateChangeConf{
		Pending: []string{"pending", "available", "deleting"},
		Target:  []string{"deleted"},
		Refresh: func() (result interface{}, state string, err error) {
			res, err := r.client.ReadNetsByIdWithResponse(ctx, data.Id.ValueString())
			if err != nil {
				response.Diagnostics.AddError("Failed to read Net on delete", err.Error())
			}

			expectedStatusCode := 200
			if res.StatusCode() != expectedStatusCode {
				// TODO: Handle NumSpot error
				apiError := utils.HandleError(res.Body)
				if apiError.Error() == "No Nets found" {
					return data, "deleted", nil
				}
				response.Diagnostics.AddError("Failed to read Net on delete", apiError.Error())
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
