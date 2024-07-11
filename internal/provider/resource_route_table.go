package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_route_table"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &RouteTableResource{}
	_ resource.ResourceWithConfigure   = &RouteTableResource{}
	_ resource.ResourceWithImportState = &RouteTableResource{}
)

type RouteTableResource struct {
	provider Provider
}

func NewRouteTableResource() resource.Resource {
	return &RouteTableResource{}
}

func (r *RouteTableResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		RouteTableFromTfToCreateRequest(&data),
		r.provider.NumspotClient.CreateRouteTableWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Route Table", err.Error())
		return
	}

	jsonRes := *res.JSON201
	createdId := *jsonRes.Id

	// Tags
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.NumspotClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	routes := make([]resource_route_table.RoutesValue, 0, len(data.Routes.Elements()))
	data.Routes.ElementsAs(ctx, &routes, false)
	diags := r.createRoutes(ctx, createdId, routes)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !data.SubnetId.IsNull() {
		diags = r.linkRouteTable(ctx, createdId, data.SubnetId.ValueString())
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	read := r.readRouteTable(ctx, createdId, response.Diagnostics)
	if read == nil {
		return
	}

	tf, diags := RouteTableFromHttpToTf(
		ctx,
		read.JSON200,
	)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
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

	tf, diags := RouteTableFromHttpToTf(
		ctx,
		res.JSON200,
	)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *RouteTableResource) readRouteTable(ctx context.Context, id string, diag diag.Diagnostics) *numspot.ReadRouteTablesByIdResponse {
	res, err := r.provider.NumspotClient.ReadRouteTablesByIdWithResponse(ctx, r.provider.SpaceID, id)
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
		state, plan resource_route_table.RouteTableModel
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
			r.provider.NumspotClient,
			r.provider.SpaceID,
			state.Id.ValueString(),
		)
		if response.Diagnostics.HasError() {
			return
		}

		modifs = true
	}

	if !state.SubnetId.Equal(plan.SubnetId) {
		currentSubnetId := state.SubnetId.ValueStringPointer()
		desiredSubnetId := plan.SubnetId.ValueStringPointer()

		switch {
		case currentSubnetId != nil && desiredSubnetId == nil:
			// Nothing to do, the subnet is deleted first, and the route table assoc is deleted
			diags = r.unlinkSubnet(ctx, state, *currentSubnetId)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
		case currentSubnetId == nil && desiredSubnetId != nil:
			// Attach create subnet
			diags = r.linkRouteTable(ctx, state.Id.ValueString(), *desiredSubnetId)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
		default:
			// Detach current and attach desired
			diags = r.unlinkSubnet(ctx, state, *currentSubnetId)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			diags = r.linkRouteTable(ctx, state.Id.ValueString(), *desiredSubnetId)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
		}
	}

	stateRoutes := make([]resource_route_table.RoutesValue, 0, len(state.Routes.Elements()))
	diags = state.Routes.ElementsAs(ctx, &stateRoutes, false)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	stateRoutesWithoutLocal := r.removeLocalRouteFromRoutes(stateRoutes)
	if !state.Routes.Equal(plan.Routes) {
		switch {
		case len(stateRoutesWithoutLocal) == 0 && len(plan.Routes.Elements()) > 0:
			planRoutes := make([]resource_route_table.RoutesValue, 0, len(plan.Routes.Elements()))
			diags = plan.Routes.ElementsAs(ctx, &planRoutes, false)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			diags = r.createRoutes(ctx, state.Id.ValueString(), planRoutes)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			modifs = true
		case len(stateRoutesWithoutLocal) > 0 && len(plan.Routes.Elements()) == 0:
			diags = r.deleteRoutes(ctx, state.Id.ValueString(), stateRoutesWithoutLocal)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			modifs = true
		case len(stateRoutesWithoutLocal) > 0 && len(plan.Routes.Elements()) > 0:
			planRoutes := make([]resource_route_table.RoutesValue, 0, len(plan.Routes.Elements()))
			diags = plan.Routes.ElementsAs(ctx, &planRoutes, false)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			toCreate, toDelete := utils.Diff(stateRoutesWithoutLocal, planRoutes)
			diags = r.createRoutes(ctx, state.Id.ValueString(), toCreate)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			diags = r.deleteRoutes(ctx, state.Id.ValueString(), toDelete)
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

		tf, diags := RouteTableFromHttpToTf(
			ctx,
			res.JSON200,
		)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
	}
}

func (r *RouteTableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_route_table.RouteTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	// Unlink route tables
	links := make([]resource_route_table.LinkRouteTablesValue, 0, len(data.LinkRouteTables.Elements()))
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

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.NumspotClient.DeleteRouteTableWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Route Table", err.Error())
		return
	}
}

func (r *RouteTableResource) createRoutes(ctx context.Context, routeTableId string, routes []resource_route_table.RoutesValue) (diags diag.Diagnostics) {
	for i := range routes {
		route := &routes[i]
		if route != nil {
			// prevent creating the one added in the plan modify function
			if !route.IsUnknown() && !route.IsNull() && !strings.EqualFold(route.GatewayId.ValueString(), "local") {
				createdRoute := utils.ExecuteRequest(func() (*numspot.CreateRouteResponse, error) {
					return r.provider.NumspotClient.CreateRouteWithResponse(ctx, r.provider.SpaceID, routeTableId, RouteTableFromTfToCreateRoutesRequest(*route))
				}, http.StatusCreated, &diags)
				if createdRoute == nil {
					return
				}
			}
		}
	}

	return
}

func (r *RouteTableResource) deleteRoutes(ctx context.Context, routeTableId string, routes []resource_route_table.RoutesValue) (diags diag.Diagnostics) {
	for i := range routes {
		route := &routes[i]
		if route != nil {
			deletedRoute := utils.ExecuteRequest(func() (*numspot.DeleteRouteResponse, error) {
				return r.provider.NumspotClient.DeleteRouteWithResponse(ctx, r.provider.SpaceID, routeTableId, RouteTableFromTfToDeleteRoutesRequest(*route))
			}, http.StatusNoContent, &diags)
			if deletedRoute == nil {
				return
			}
		}
	}

	return
}

func (r *RouteTableResource) unlinkSubnet(ctx context.Context, state resource_route_table.RouteTableModel, subnetId string) diag.Diagnostics {
	// This function attempts to unlink a subnet from a route table.
	// It first retrieves all linked route tables, then identifies the specific association by subnet ID.
	// If found, it proceeds to unlink the route table using the association ID.
	// Any errors encountered during the process are added to the diagnostics and returned.

	var diags diag.Diagnostics
	rtbAssocsTf := make([]resource_route_table.LinkRouteTablesValue, 0, len(state.LinkRouteTables.Elements()))
	diags = state.LinkRouteTables.ElementsAs(ctx, &rtbAssocsTf, false)
	if diags.HasError() {
		return diags
	}

	var rtbAssocId *string
	for _, rtbAssocTf := range rtbAssocsTf {
		if strings.EqualFold(rtbAssocTf.SubnetId.ValueString(), subnetId) {
			rtbAssocId = rtbAssocTf.Id.ValueStringPointer()
			break
		}
	}

	if rtbAssocId == nil {
		diags.AddError("Failed to retrieve route table associated with subnet Id", "Failed to retrieve route table associated with subnet Id")
		return diags
	}

	unlinkDiags := r.unlinkRouteTable(ctx, state.Id.ValueString(), *rtbAssocId)

	// This check is necessary to ensure the operation is successful in scenarios where the subnet was created prior to this action.
	// In such cases, the association link might be deleted before the route table resource has been updated, leading to potential inconsistencies.
	if unlinkDiags.HasError() {
		readRes := r.readRouteTable(ctx, state.Id.ValueString(), diags)
		if readRes == nil {
			return diags
		}

		found := false
		for _, rtb := range *readRes.JSON200.LinkRouteTables {
			if rtb.Id == rtbAssocId {
				found = true
				break
			}
		}

		if found {
			diags.Append(unlinkDiags...)
			return diags
		}
	}

	return diags
}

func (r *RouteTableResource) linkRouteTable(ctx context.Context, routeTableId, subnetId string) diag.Diagnostics {
	var diags diag.Diagnostics

	utils.ExecuteRequest(func() (*numspot.LinkRouteTableResponse, error) {
		return r.provider.NumspotClient.LinkRouteTableWithResponse(ctx, r.provider.SpaceID, routeTableId, numspot.LinkRouteTableJSONRequestBody{SubnetId: subnetId})
	}, http.StatusOK, &diags)

	return diags
}

func (r *RouteTableResource) unlinkRouteTable(ctx context.Context, routeTableId, linkRouteTableId string) diag.Diagnostics {
	var diags diag.Diagnostics

	utils.ExecuteRequest(func() (*numspot.UnlinkRouteTableResponse, error) {
		return r.provider.NumspotClient.UnlinkRouteTableWithResponse(ctx, r.provider.SpaceID, routeTableId, numspot.UnlinkRouteTableJSONRequestBody{LinkRouteTableId: linkRouteTableId})
	}, http.StatusNoContent, &diags)

	return diags
}

func (r *RouteTableResource) removeLocalRouteFromRoutes(routes []resource_route_table.RoutesValue) []resource_route_table.RoutesValue {
	arr := make([]resource_route_table.RoutesValue, 0)
	for _, route := range routes {
		if !strings.EqualFold(route.GatewayId.ValueString(), "local") {
			arr = append(arr, route)
		}
	}

	return arr
}
