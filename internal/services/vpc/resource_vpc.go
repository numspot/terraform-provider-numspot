package vpc

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
	_ resource.Resource                = &VpcResource{}
	_ resource.ResourceWithConfigure   = &VpcResource{}
	_ resource.ResourceWithImportState = &VpcResource{}
)

type VpcResource struct {
	provider services.IProvider
}

func NewNetResource() resource.Resource {
	return &VpcResource{}
}

func (r *VpcResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *VpcResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VpcResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpc"
}

func (r *VpcResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VpcResourceSchema(ctx)
}

func (r *VpcResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data VpcModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils2.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		NetFromTfToCreateRequest(&data),
		r.provider.GetNumspotClient().CreateVpcWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPC", err.Error())
		return
	}

	// Handle tags
	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// if dhcp_options_set_id is set, we need to update the Vpc as this attribute can be set on Update only and not on Create
	if !data.DhcpOptionsSetId.IsNull() && !data.DhcpOptionsSetId.IsUnknown() {
		updatedRes := utils2.ExecuteRequest(func() (*numspot.UpdateVpcResponse, error) {
			body := VpcFromTfToUpdaterequest(ctx, &data, &response.Diagnostics)
			return r.provider.GetNumspotClient().UpdateVpcWithResponse(ctx, r.provider.GetSpaceID(), createdId, body)
		}, http.StatusOK, &response.Diagnostics)

		if updatedRes == nil || response.Diagnostics.HasError() {
			return
		}
	}
	readRes, err := utils2.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.GetSpaceID(),
		[]string{"pending"},
		[]string{"available"},
		r.provider.GetNumspotClient().ReadVpcsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Net", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", createdId, err))
		return
	}

	vpc, ok := readRes.(*numspot.Vpc)
	if !ok {
		response.Diagnostics.AddError("Failed to read VPC", "object conversion error")
		return
	}

	tf, diags := NetFromHttpToTf(ctx, vpc)
	if diags.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *VpcResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data VpcModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils2.ExecuteRequest(func() (*numspot.ReadVpcsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadVpcsByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	// TODO: read Nets returns tags in response, do not need to relist tags
	tf, diags := NetFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		return
	}

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *VpcResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state VpcModel
		plan  VpcModel
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vpcId := state.Id.ValueString()

	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			vpcId,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Update Vpc
	updatedRes := utils2.ExecuteRequest(func() (*numspot.UpdateVpcResponse, error) {
		body := VpcFromTfToUpdaterequest(ctx, &plan, &response.Diagnostics)
		return r.provider.GetNumspotClient().UpdateVpcWithResponse(ctx, r.provider.GetSpaceID(), vpcId, body)
	}, http.StatusOK, &response.Diagnostics)

	if updatedRes == nil || response.Diagnostics.HasError() {
		return
	}

	// Read resource
	res := utils2.ExecuteRequest(func() (*numspot.ReadVpcsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadVpcsByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := NetFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		return
	}
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *VpcResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VpcModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils2.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteVpcWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete VPC", err.Error())
		return
	}

	response.State.RemoveResource(ctx)
}
