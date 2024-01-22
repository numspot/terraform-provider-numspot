package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_route_table"
)

var (
	_ resource.Resource                = &RouteTableResource{}
	_ resource.ResourceWithConfigure   = &RouteTableResource{}
	_ resource.ResourceWithImportState = &RouteTableResource{}
)

type RouteTableResource struct {
	client *api.ClientWithResponses
}

func NewRouteTableResource() resource.Resource {
	return &RouteTableResource{}
}

func (r *RouteTableResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *RouteTableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *RouteTableResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_route_table"
}

func (r *RouteTableResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_route_table.RouteTableResourceSchema(ctx)
}

func (r *RouteTableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_route_table.RouteTableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	body := RouteTableFromTfToCreateRequest(data)
	res, err := r.client.CreateRouteTableWithResponse(ctx, body)
	if err != nil {
		response.Diagnostics.AddError("Failed to create RouteTable", err.Error())
		return
	}

	expectedStatusCode := 200
	if res.StatusCode() != expectedStatusCode {
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to create RouteTable", apiError.Error())
		return
	}

	createdId := res.JSON200.Id

	// Delete default

	routes := make([]resource_route_table.RoutesValue, 0, len(data.Routes.Elements()))
	data.Routes.ElementsAs(ctx, &routes, false)
	for _, route := range routes {
		createRouteRes, err := r.client.CreateRouteWithResponse(ctx, api.CreateRouteJSONRequestBody{
			DestinationIpRange: route.DestinationIpRange.ValueString(),
			GatewayId:          route.GatewayId.ValueStringPointer(),
			NatServiceId:       route.NatServiceId.ValueStringPointer(),
			NetPeeringId:       route.NetPeeringId.ValueStringPointer(),
			NicId:              route.NicId.ValueStringPointer(),
			TableId:            createdId,
			VmId:               route.VmId.ValueStringPointer(),
		})

		if err != nil {
			response.Diagnostics.AddError("Failed to create RouteTable route", err.Error())
			return
		}

		if createRouteRes.StatusCode() != 200 {
			apiError := utils.HandleError(res.Body)
			response.Diagnostics.AddError("Failed to create RouteTable route", apiError.Error())
			return
		}
	}

	readed := r.readRouteTable(ctx, *createdId, response.Diagnostics)
	if readed == nil {
		return
	}

	tf, diag := RouteTableFromHttpToTf(ctx, readed.JSON200)
	if diag.HasError() {
		response.Diagnostics.Append(diag...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *RouteTableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_route_table.RouteTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := r.readRouteTable(ctx, data.Id.ValueString(), response.Diagnostics)
	if res == nil {
		return
	}

	tf, diag := RouteTableFromHttpToTf(ctx, res.JSON200) // FIXME
	if diag.HasError() {
		response.Diagnostics.Append(diag...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *RouteTableResource) readRouteTable(ctx context.Context, id string, diag diag.Diagnostics) *api.ReadRouteTablesByIdResponse {
	res, err := r.client.ReadRouteTablesByIdWithResponse(ctx, id)
	if err != nil {
		diag.AddError("Failed to read RouteTable", err.Error())
		return nil
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
	if res.StatusCode() != expectedStatusCode {
		apiError := utils.HandleError(res.Body)
		diag.AddError("Failed to read RouteTable", apiError.Error())
		return nil
	}

	return res
}

func (r *RouteTableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *RouteTableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_route_table.RouteTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// TODO: Implement DELETE operation
	res, err := r.client.DeleteRouteTableWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete RouteTable", err.Error())
		return
	}

	expectedStatusCode := 200 // FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to delete RouteTable", apiError.Error())
		return
	}
}
