package vpcpeering

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource              = &VpcPeeringResource{}
	_ resource.ResourceWithConfigure = &VpcPeeringResource{}
)

type VpcPeeringResource struct {
	provider services.IProvider
}

func NewVpcPeeringResource() resource.Resource {
	return &VpcPeeringResource{}
}

func (r *VpcPeeringResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *VpcPeeringResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpc_peering"
}

func (r *VpcPeeringResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VpcPeeringResourceSchema(ctx)
}

func (r *VpcPeeringResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data VpcPeeringModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils2.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		VpcPeeringFromTfToCreateRequest(data),
		r.provider.GetNumspotClient().CreateVpcPeeringWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPC Peering", err.Error())
		return
	}

	// Retries can't success with retry_utils.RetryReadUntilStateValid because vpc_peering state is an object but a string.
	// Also, VPC Peering resource State is used to provided status of the peering process, not the creation process
	// So we do not need to implement specific retry process here.

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	readRes := utils2.ExecuteRequest(func() (*numspot.ReadVpcPeeringsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadVpcPeeringsByIdWithResponse(ctx, r.provider.GetSpaceID(), createdId)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := VpcPeeringFromHttpToTf(ctx, readRes.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	if tf.SourceVpcId.IsNull() {
		tf.SourceVpcId = data.SourceVpcId
	}

	if tf.AccepterVpcId.IsNull() {
		tf.AccepterVpcId = data.AccepterVpcId
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VpcPeeringResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data VpcPeeringModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils2.ExecuteRequest(func() (*numspot.ReadVpcPeeringsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadVpcPeeringsByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := VpcPeeringFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	if tf.SourceVpcId.IsNull() {
		tf.SourceVpcId = data.SourceVpcId
	}

	if tf.AccepterVpcId.IsNull() {
		tf.AccepterVpcId = data.AccepterVpcId
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VpcPeeringResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan VpcPeeringModel
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

	res := utils2.ExecuteRequest(func() (*numspot.ReadVpcPeeringsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadVpcPeeringsByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := VpcPeeringFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	if tf.SourceVpcId.IsNull() {
		tf.SourceVpcId = state.SourceVpcId
	}

	if tf.AccepterVpcId.IsNull() {
		tf.AccepterVpcId = state.AccepterVpcId
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VpcPeeringResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VpcPeeringModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils2.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteVpcPeeringWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete VPC Peering", err.Error())
		return
	}
}
