package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_internet_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &InternetGatewayResource{}
	_ resource.ResourceWithConfigure   = &InternetGatewayResource{}
	_ resource.ResourceWithImportState = &InternetGatewayResource{}
)

type InternetGatewayResource struct {
	provider Provider
}

func NewInternetGatewayResource() resource.Resource {
	return &InternetGatewayResource{}
}

func (r *InternetGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *InternetGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *InternetGatewayResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_internet_gateway"
}

func (r *InternetGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_internet_gateway.InternetGatewayResourceSchema(ctx)
}

func (r *InternetGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailable(
		ctx,
		r.provider.SpaceID,
		r.provider.NumspotClient.CreateInternetGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Internet Gateway", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.NumspotClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Call Link Internet Service to VPC
	vpcId := data.VpcIp
	if !vpcId.IsNull() {
		linRes := utils.ExecuteRequest(func() (*numspot.LinkInternetGatewayResponse, error) {
			return r.provider.NumspotClient.LinkInternetGatewayWithResponse(
				ctx,
				r.provider.SpaceID,
				createdId,
				numspot.LinkInternetGatewayJSONRequestBody{
					VpcId: data.VpcIp.ValueString(),
				},
			)
		}, http.StatusNoContent, &response.Diagnostics)
		if linRes == nil {
			return
		}
	}

	read, err := retry_utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{},
		[]string{"available"},
		r.provider.NumspotClient.ReadInternetGatewaysByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Internet Gateway", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	rr, ok := read.(*numspot.InternetGateway)
	if !ok {
		response.Diagnostics.AddError("Failed to create internet gateway", "object conversion error")
		return
	}

	tf, diags := InternetServiceFromHttpToTf(ctx, rr)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	tf.VpcIp = data.VpcIp
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *InternetGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*numspot.ReadInternetGatewaysByIdResponse, error) {
		return r.provider.NumspotClient.ReadInternetGatewaysByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := InternetServiceFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *InternetGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan resource_internet_gateway.InternetGatewayModel
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
			r.provider.NumspotClient,
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

	res := utils.ExecuteRequest(func() (*numspot.ReadInternetGatewaysByIdResponse, error) {
		return r.provider.NumspotClient.ReadInternetGatewaysByIdWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := InternetServiceFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *InternetGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	tflog.Debug(ctx, fmt.Sprintf("Deleting internet gateway: %s", data.Id.ValueString()))

	if !data.VpcIp.IsNull() {
		tflog.Debug(ctx, fmt.Sprintf("Detaching vpc: %s, from internet gateway: %s", data.VpcIp.ValueString(), data.Id.ValueString()))

		err := retry_utils.RetryUnlinkUntilSuccess(
			ctx,
			r.provider.SpaceID,
			data.Id.ValueString(),
			numspot.UnlinkInternetGatewayJSONRequestBody{
				VpcId: data.VpcIp.ValueString(),
			},
			r.provider.NumspotClient.UnlinkInternetGatewayWithResponse,
		)
		if err != nil {
			response.Diagnostics.AddError("Failed to delete Internet Gateway", err.Error())
			return
		}
	}

	err := retry_utils.RetryDeleteUntilResourceAvailable(
		ctx,
		r.provider.SpaceID,
		data.Id.ValueString(),
		r.provider.NumspotClient.DeleteInternetGatewayWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Internet Gateway", err.Error())
		return
	}
}
