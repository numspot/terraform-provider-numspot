package internet_gateway

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/common/slice"

	api_client "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api_client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns"
)

var (
	_ resource.Resource                = &InternetGatewayResource{}
	_ resource.ResourceWithConfigure   = &InternetGatewayResource{}
	_ resource.ResourceWithImportState = &InternetGatewayResource{}
)

func NewInternetGatewayResource() resource.Resource {
	return &InternetGatewayResource{}
}

type InternetGatewayResource struct {
	client *api_client.ClientWithResponses
}

type InternetGatewayResourceModel struct {
	Id    types.String `tfsdk:"id"`
	VpcId types.String `tfsdk:"virtual_private_cloud_id"`
}

func (k *InternetGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "NumSpot internet gateway resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The NumSpot internet gateway resource computed id.",
				Computed:            true,
			},
			"virtual_private_cloud_id": schema.StringAttribute{
				MarkdownDescription: "The id of the VPC to connect to.",
				Optional:            true,
			},
		},
	}
}

func (k *InternetGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (k *InternetGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client := conns.GetClient(request, response)
	if client == nil || response.Diagnostics.HasError() {
		return
	}
	k.client = client
}

func (k *InternetGatewayResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_internet_gateway"
}

func (k *InternetGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan InternetGatewayResourceModel

	// Read plan
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Create non-connected internet gateway
	createInternetGatewayResponse, err := k.client.CreateInternetGatewayWithResponse(ctx)
	if err != nil {
		response.Diagnostics.AddError("Creating Internet Gateway", err.Error())
		return
	}
	numspotError := conns.HandleError(http.StatusCreated,
		createInternetGatewayResponse.HTTPResponse.StatusCode,
		createInternetGatewayResponse.Body,
	)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}

	if !plan.VpcId.IsNull() {

		// Attach internet gateway to the vpc
		err, numspotError = k.attachInternetGateway(ctx, *createInternetGatewayResponse.JSON201.Id, plan.VpcId.ValueString())
		if err != nil {
			response.Diagnostics.AddError("Attaching Internet Gateway", err.Error())
			return
		}
		if numspotError != nil {
			response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
			return
		}
	}

	data := InternetGatewayResourceModel{
		Id:    types.StringValue(*createInternetGatewayResponse.JSON201.Id),
		VpcId: plan.VpcId,
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (k *InternetGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state InternetGatewayResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if plan.VpcId.ValueString() != state.VpcId.ValueString() {
		// Detach internet gateway from the current vpc
		err, numspotError := k.detachInternetGateway(ctx, state)
		if err != nil {
			response.Diagnostics.AddError("Detaching Internet Gateway", err.Error())
			return
		}
		if numspotError != nil {
			response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
			return
		}

		// Attach internet gateway to the new vpc
		err, numspotError = k.attachInternetGateway(ctx, state.Id.ValueString(), plan.VpcId.ValueString())
		if err != nil {
			response.Diagnostics.AddError("Attaching Internet Gateway", err.Error())
			return
		}
		if numspotError != nil {
			response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
			return
		}

		data := InternetGatewayResourceModel{
			Id:    state.Id,
			VpcId: plan.VpcId,
		}

		response.Diagnostics.Append(response.State.Set(ctx, &data)...)
		if response.Diagnostics.HasError() {
			return
		}
	}
	response.Diagnostics.Append(response.State.Get(ctx, &plan)...)
}

func (k *InternetGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data InternetGatewayResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	internetGateways, err := k.client.ListInternetGatewayWithResponse(ctx)
	if err != nil {
		response.Diagnostics.AddError("Reading Internet Gateways", err.Error())
		return
	}

	internetGateway := slice.FindFirst(
		*internetGateways.JSON200.Items,
		func(e api_client.InternetGateway) bool {
			return *e.Id == data.Id.ValueString()
		},
	)

	if internetGateway == nil {
		response.State.RemoveResource(ctx)
	}

	response.Diagnostics.Append(response.State.Set(ctx, mapToInternetGatewayModel(internetGateway))...)
}

func mapToInternetGatewayModel(internetGateway *api_client.InternetGateway) InternetGatewayResourceModel {
	var vpcID basetypes.StringValue

	switch {
	case internetGateway.VirtualPrivateCloudId == nil:
		vpcID = types.StringNull()
	default:
		vpcID = types.StringValue(*internetGateway.VirtualPrivateCloudId)
	}

	return InternetGatewayResourceModel{
		Id:    types.StringValue(*internetGateway.Id),
		VpcId: vpcID,
	}
}

func (k *InternetGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state InternetGatewayResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	if !state.VpcId.IsNull() {
		err, numspotError := k.detachInternetGateway(ctx, state)
		if err != nil {
			response.Diagnostics.AddError("Detaching Internet Gateway", err.Error())
			return
		}
		if numspotError != nil {
			response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
			return
		}
	}

	res, err := k.client.DeleteInternetGatewayWithResponse(ctx, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Deleting Internet Gateway", err.Error())
		return
	}

	numspotError := conns.HandleError(http.StatusNoContent, res.HTTPResponse.StatusCode, res.Body)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}
}

func (k *InternetGatewayResource) attachInternetGateway(ctx context.Context, internetGatewayID, vpcID string) (error, *api_client.Error) {
	body := api_client.AttachInternetGatewayJSONRequestBody{
		VirtualPrivateCloudId: vpcID,
	}

	attachInternetGatewayResponse, err := k.client.AttachInternetGatewayWithResponse(
		ctx, internetGatewayID, body)
	if err != nil {
		return err, nil
	}

	numspotError := conns.HandleError(http.StatusNoContent,
		attachInternetGatewayResponse.HTTPResponse.StatusCode,
		attachInternetGatewayResponse.Body,
	)
	if numspotError != nil {
		return nil, numspotError
	}
	return nil, nil
}

func (k *InternetGatewayResource) detachInternetGateway(ctx context.Context, internetGateway InternetGatewayResourceModel) (error, *api_client.Error) {
	body := api_client.DetachInternetGatewayJSONRequestBody{
		VirtualPrivateCloudId: internetGateway.VpcId.ValueString(),
	}

	detachInternetGatewayResponse, err := k.client.DetachInternetGatewayWithResponse(
		ctx, internetGateway.Id.ValueString(), body)
	if err != nil {
		return err, nil
	}

	numspotError := conns.HandleError(http.StatusNoContent,
		detachInternetGatewayResponse.HTTPResponse.StatusCode,
		detachInternetGatewayResponse.Body,
	)
	if numspotError != nil {
		return nil, numspotError
	}
	return nil, nil
}
