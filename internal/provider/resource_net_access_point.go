package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_net_access_point"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &NetAccessPointResource{}
	_ resource.ResourceWithConfigure   = &NetAccessPointResource{}
	_ resource.ResourceWithImportState = &NetAccessPointResource{}
)

type NetAccessPointResource struct {
	provider Provider
}

func NewNetAccessPointResource() resource.Resource {
	return &NetAccessPointResource{}
}

func (r *NetAccessPointResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *NetAccessPointResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *NetAccessPointResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_net_access_point"
}

func (r *NetAccessPointResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_net_access_point.NetAccessPointResourceSchema(ctx)
}

func (r *NetAccessPointResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_net_access_point.NetAccessPointModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		NetAccessPointFromTfToCreateRequest(ctx, &data),
		r.provider.ApiClient.CreateVpcAccessPointWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPC Access Point", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.ApiClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Retries read on resource until state is OK
	read, err := retry_utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"available"},
		r.provider.ApiClient.ReadVpcAccessPointsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPC Access Point", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	rr, ok := read.(*iaas.VpcAccessPoint)
	if !ok {
		response.Diagnostics.AddError("Failed to create vpc access point", "object conversion error")
		return
	}

	tf, diags := NetAccessPointFromHttpToTf(ctx, rr)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NetAccessPointResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_net_access_point.NetAccessPointModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadVpcAccessPointsByIdResponse, error) {
		return r.provider.ApiClient.ReadVpcAccessPointsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := NetAccessPointFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NetAccessPointResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan resource_net_access_point.NetAccessPointModel
	modifications := false
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

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
		modifications = true
	}

	if !modifications {
		return
	}

	res := utils.ExecuteRequest(func() (*iaas.ReadVpcAccessPointsByIdResponse, error) {
		return r.provider.ApiClient.ReadVpcAccessPointsByIdWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := NetAccessPointFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *NetAccessPointResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_net_access_point.NetAccessPointModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	_ = utils.ExecuteRequest(func() (*iaas.DeleteVpcAccessPointResponse, error) {
		return r.provider.ApiClient.DeleteVpcAccessPointWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
}
