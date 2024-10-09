package routetable

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &RouteTableResource{}
	_ resource.ResourceWithConfigure   = &RouteTableResource{}
	_ resource.ResourceWithImportState = &RouteTableResource{}
)

type RouteTableResource struct {
	provider services.IProvider
}

func NewRouteTableResource() resource.Resource {
	return &RouteTableResource{}
}

func (r *RouteTableResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *RouteTableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *RouteTableResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_route_table"
}

func (r *RouteTableResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = RouteTableResourceSchema(ctx)
}

func (r *RouteTableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data RouteTableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		RouteTableFromTfToCreateRequest(&data),
		r.provider.GetNumspotClient().CreateRouteTableWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Route Table", err.Error())
		return
	}

	jsonRes := *res.JSON201
	createdId := *jsonRes.Id

	// Tags
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	routes := make([]RoutesValue, 0, len(data.Routes.Elements()))
	data.Routes.ElementsAs(ctx, &routes, false)
	r.createRoutes(ctx, createdId, routes, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	if !data.SubnetId.IsNull() {
		r.linkRouteTable(ctx, createdId, data.SubnetId.ValueString(), &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	read := r.readRouteTable(ctx, createdId, response.Diagnostics)
	if read == nil {
		return
	}

	tf := RouteTableFromHttpToTf(
		ctx,
		read.JSON200,
		&response.Diagnostics,
	)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *RouteTableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data RouteTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := r.readRouteTable(ctx, data.Id.ValueString(), response.Diagnostics)
	if res == nil {
		return
	}

	tf := RouteTableFromHttpToTf(
		ctx,
		res.JSON200,
		&response.Diagnostics,
	)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *RouteTableResource) readRouteTable(ctx context.Context, id string, diag diag.Diagnostics) *numspot.ReadRouteTablesByIdResponse {
	res, err := r.provider.GetNumspotClient().ReadRouteTablesByIdWithResponse(ctx, r.provider.GetSpaceID(), id)
	if err != nil {
		diag.AddError("Failed to read RouteTable", err.Error())
		return nil
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		diag.AddError("Failed to read RouteTable", apiError.Error())
		return nil
	}

	return res
}

func (r *RouteTableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan RouteTableModel
		modifs      = false
	)

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	diags = request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			state.Id.ValueString(),
		)
		if response.Diagnostics.HasError() {
			return
		}

		modifs = true
	}

	stateRoutes := make([]RoutesValue, 0, len(state.Routes.Elements()))
	diags = state.Routes.ElementsAs(ctx, &stateRoutes, false)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	stateRoutesWithoutLocal := r.removeLocalRouteFromRoutes(stateRoutes)
	if !state.Routes.Equal(plan.Routes) {
		switch {
		case len(stateRoutesWithoutLocal) == 0 && len(plan.Routes.Elements()) > 0:
			planRoutes := make([]RoutesValue, 0, len(plan.Routes.Elements()))
			diags = plan.Routes.ElementsAs(ctx, &planRoutes, false)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			r.createRoutes(ctx, state.Id.ValueString(), planRoutes, &response.Diagnostics)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			modifs = true
		case len(stateRoutesWithoutLocal) > 0 && len(plan.Routes.Elements()) == 0:
			r.deleteRoutes(ctx, state.Id.ValueString(), stateRoutesWithoutLocal, &response.Diagnostics)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			modifs = true
		case len(stateRoutesWithoutLocal) > 0 && len(plan.Routes.Elements()) > 0:
			planRoutes := make([]RoutesValue, 0, len(plan.Routes.Elements()))
			diags = plan.Routes.ElementsAs(ctx, &planRoutes, false)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			toCreate, toDelete := utils.Diff(stateRoutesWithoutLocal, planRoutes)
			r.createRoutes(ctx, state.Id.ValueString(), toCreate, &response.Diagnostics)
			if response.Diagnostics.HasError() {
				return
			}

			r.deleteRoutes(ctx, state.Id.ValueString(), toDelete, &response.Diagnostics)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			modifs = true
		}
	}

	if modifs {
		// Read and update the state
		res := r.readRouteTable(ctx, state.Id.ValueString(), response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		tf := RouteTableFromHttpToTf(
			ctx,
			res.JSON200,
			&response.Diagnostics,
		)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
	}
}

func (r *RouteTableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data RouteTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// Unlink route tables
	links := make([]LinkRouteTablesValue, 0, len(data.LinkRouteTables.Elements()))
	diags := data.LinkRouteTables.ElementsAs(ctx, &links, false)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	unlinkDiags := diag.Diagnostics{}
	for _, link := range links {
		diags = r.unlinkRouteTable(ctx, data.Id.ValueString(), link.Id.ValueString())
		unlinkDiags.Append(diags...)
	}

	if unlinkDiags.HasError() {
		readRes := r.readRouteTable(ctx, data.Id.ValueString(), response.Diagnostics)
		if readRes == nil {
			return
		}

		if len(*readRes.JSON200.LinkRouteTables) > 0 {
			response.Diagnostics.Append(unlinkDiags...)
			return
		}
	}

	err := utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteRouteTableWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Route Table", err.Error())
		return
	}
}

func (r *RouteTableResource) createRoutes(ctx context.Context, routeTableId string, routes []RoutesValue, diags *diag.Diagnostics) {
	for i := range routes {
		route := &routes[i]
		if route != nil {
			// prevent creating the one added in the plan modify function
			if !route.IsUnknown() && !route.IsNull() && !strings.EqualFold(route.GatewayId.ValueString(), "local") {
				createdRoute := utils.ExecuteRequest(func() (*numspot.CreateRouteResponse, error) {
					return r.provider.GetNumspotClient().CreateRouteWithResponse(ctx, r.provider.GetSpaceID(), routeTableId, RouteTableFromTfToCreateRoutesRequest(*route))
				}, http.StatusCreated, diags)
				if createdRoute == nil {
					return
				}
			}
		}
	}
}

func (r *RouteTableResource) deleteRoutes(ctx context.Context, routeTableId string, routes []RoutesValue, diags *diag.Diagnostics) {
	for i := range routes {
		route := &routes[i]
		if route != nil {
			deletedRoute := utils.ExecuteRequest(func() (*numspot.DeleteRouteResponse, error) {
				return r.provider.GetNumspotClient().DeleteRouteWithResponse(ctx, r.provider.GetSpaceID(), routeTableId, RouteTableFromTfToDeleteRoutesRequest(*route))
			}, http.StatusNoContent, diags)
			if deletedRoute == nil {
				return
			}
		}
	}
}

func (r *RouteTableResource) linkRouteTable(ctx context.Context, routeTableId, subnetId string, diags *diag.Diagnostics) {
	utils.ExecuteRequest(func() (*numspot.LinkRouteTableResponse, error) {
		return r.provider.GetNumspotClient().LinkRouteTableWithResponse(ctx, r.provider.GetSpaceID(), routeTableId, numspot.LinkRouteTableJSONRequestBody{SubnetId: subnetId})
	}, http.StatusOK, diags)
}

func (r *RouteTableResource) unlinkRouteTable(ctx context.Context, routeTableId, linkRouteTableId string) diag.Diagnostics {
	var diags diag.Diagnostics

	utils.ExecuteRequest(func() (*numspot.UnlinkRouteTableResponse, error) {
		return r.provider.GetNumspotClient().UnlinkRouteTableWithResponse(ctx, r.provider.GetSpaceID(), routeTableId, numspot.UnlinkRouteTableJSONRequestBody{LinkRouteTableId: linkRouteTableId})
	}, http.StatusNoContent, &diags)

	return diags
}

func (r *RouteTableResource) removeLocalRouteFromRoutes(routes []RoutesValue) []RoutesValue {
	arr := make([]RoutesValue, 0)
	for _, route := range routes {
		if !strings.EqualFold(route.GatewayId.ValueString(), "local") {
			arr = append(arr, route)
		}
	}

	return arr
}
