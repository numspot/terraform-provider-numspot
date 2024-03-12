package provider

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_dhcp_options"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &DhcpOptionsResource{}
	_ resource.ResourceWithConfigure   = &DhcpOptionsResource{}
	_ resource.ResourceWithImportState = &DhcpOptionsResource{}
)

type DhcpOptionsResource struct {
	provider Provider
}

func NewDhcpOptionsResource() resource.Resource {
	return &DhcpOptionsResource{}
}

func (r *DhcpOptionsResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *DhcpOptionsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *DhcpOptionsResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_dhcp_options"
}

func (r *DhcpOptionsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_dhcp_options.DhcpOptionsResourceSchema(ctx)
}

func (r *DhcpOptionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_dhcp_options.DhcpOptionsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.CreateDhcpOptionsResponse, error) {
		body := DhcpOptionsFromTfToCreateRequest(ctx, data)
		return r.provider.ApiClient.CreateDhcpOptionsWithResponse(ctx, r.provider.SpaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	createdId := *res.JSON201.Id
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.ApiClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	tf, diagnostics := DhcpOptionsFromHttpToTf(ctx, res.JSON201)
	tf.Tags = tags.ReadTags(ctx, r.provider.ApiClient, r.provider.SpaceID, response.Diagnostics, createdId)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DhcpOptionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_dhcp_options.DhcpOptionsModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadDhcpOptionsByIdResponse, error) {
		return r.provider.ApiClient.ReadDhcpOptionsByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := DhcpOptionsFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DhcpOptionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resource_dhcp_options.DhcpOptionsModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

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
	}

	res := utils.ExecuteRequest(func() (*api.ReadDhcpOptionsByIdResponse, error) {
		return r.provider.ApiClient.ReadDhcpOptionsByIdWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := DhcpOptionsFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		return
	}
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *DhcpOptionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_dhcp_options.DhcpOptionsModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.DeleteDhcpOptionsResponse, error) {
		return r.provider.ApiClient.DeleteDhcpOptionsWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}
}
