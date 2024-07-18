package subnet

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &SubnetResource{}
	_ resource.ResourceWithConfigure   = &SubnetResource{}
	_ resource.ResourceWithImportState = &SubnetResource{}
)

type SubnetResource struct {
	provider services.IProvider
}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

func (r *SubnetResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *SubnetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *SubnetResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = SubnetResourceSchema(ctx)
}

func (r *SubnetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data SubnetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils2.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		SubnetFromTfToCreateRequest(&data),
		r.provider.GetNumspotClient().CreateSubnetWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Subnet", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	_, err = utils2.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.GetSpaceID(),
		[]string{"pending"},
		[]string{"available"},
		r.provider.GetNumspotClient().ReadSubnetsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Net", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", *res.JSON201.Id, err))
		return
	}

	if data.MapPublicIpOnLaunch.ValueBool() {
		updateRes := utils2.ExecuteRequest(func() (*numspot.UpdateSubnetResponse, error) {
			return r.provider.GetNumspotClient().UpdateSubnetWithResponse(ctx, r.provider.GetSpaceID(), createdId, numspot.UpdateSubnetJSONRequestBody{
				MapPublicIpOnLaunch: true,
			})
		}, http.StatusOK, &response.Diagnostics)
		if updateRes == nil {
			return
		}
	}

	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	readRes := utils2.ExecuteRequest(func() (*numspot.ReadSubnetsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadSubnetsByIdWithResponse(ctx, r.provider.GetSpaceID(), createdId)
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
	var data SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils2.ExecuteRequest(func() (*numspot.ReadSubnetsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadSubnetsByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
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
	var plan, state SubnetModel

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
	}

	res := utils2.ExecuteRequest(func() (*numspot.ReadSubnetsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadSubnetsByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
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
	var data SubnetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils2.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteSubnetWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete subnet", err.Error())
		return
	}

	response.State.RemoveResource(ctx)
}
