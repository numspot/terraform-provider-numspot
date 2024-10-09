package dhcpoptions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &DhcpOptionsResource{}
	_ resource.ResourceWithConfigure   = &DhcpOptionsResource{}
	_ resource.ResourceWithImportState = &DhcpOptionsResource{}
)

type DhcpOptionsResource struct {
	provider services.IProvider
}

func NewDhcpOptionsResource() resource.Resource {
	return &DhcpOptionsResource{}
}

func (r *DhcpOptionsResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *DhcpOptionsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *DhcpOptionsResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_dhcp_options"
}

func (r *DhcpOptionsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = DhcpOptionsResourceSchema(ctx)
}

func (r *DhcpOptionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data DhcpOptionsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		DhcpOptionsFromTfToCreateRequest(ctx, data),
		r.provider.GetNumspotClient().CreateDhcpOptionsWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create DHCP Options", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	tf := DhcpOptionsFromHttpToTf(ctx, res.JSON201, &response.Diagnostics)
	tf.Tags = tags.ReadTags(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), response.Diagnostics, createdId)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DhcpOptionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data DhcpOptionsModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*numspot.ReadDhcpOptionsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadDhcpOptionsByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := DhcpOptionsFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DhcpOptionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state DhcpOptionsModel

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

	res := utils.ExecuteRequest(func() (*numspot.ReadDhcpOptionsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadDhcpOptionsByIdWithResponse(ctx, r.provider.GetSpaceID(), state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := DhcpOptionsFromHttpToTf(ctx, res.JSON200, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DhcpOptionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data DhcpOptionsModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteDhcpOptionsWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete DHCP Options", err.Error())
		return
	}
}
