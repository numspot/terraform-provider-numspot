package netaccesspoint

/*

Net Access Points are not handled for now


import (
	"context"
	"fmt"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &NetAccessPointResource{}
	_ resource.ResourceWithConfigure   = &NetAccessPointResource{}
	_ resource.ResourceWithImportState = &NetAccessPointResource{}
)

type NetAccessPointResource struct {
	provider services.IProvider
}

func NewNetAccessPointResource() resource.Resource {
	return &NetAccessPointResource{}
}

func (r *NetAccessPointResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *NetAccessPointResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *NetAccessPointResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_net_access_point"
}

func (r *NetAccessPointResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = NetAccessPointResourceSchema(ctx)
}

func (r *NetAccessPointResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data NetAccessPointModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		NetAccessPointFromTfToCreateRequest(ctx, &data),
		r.provider.GetNumspotClient().CreateVpcAccessPointWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPC Access Point", err.Error())
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
	read, err := utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.GetSpaceID(),
		[]string{"pending"},
		[]string{"available"},
		r.provider.GetNumspotClient().ReadVpcAccessPointsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPC Access Point", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	rr, ok := read.(*numspot.VpcAccessPoint)
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
	var data NetAccessPointModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*numspot.ReadVpcAccessPointsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadVpcAccessPointsByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
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
	var state, plan NetAccessPointModel
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
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
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

	res := utils.ExecuteRequest(func() (*numspot.ReadVpcAccessPointsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadVpcAccessPointsByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
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
	var data NetAccessPointModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	_ = utils.ExecuteRequest(func() (*numspot.DeleteVpcAccessPointResponse, error) {
		return r.provider.GetNumspotClient().DeleteVpcAccessPointWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
}
*/
