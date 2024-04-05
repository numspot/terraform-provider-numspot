package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_listener_rule"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &ListenerRuleResource{}
	_ resource.ResourceWithConfigure   = &ListenerRuleResource{}
	_ resource.ResourceWithImportState = &ListenerRuleResource{}
)

type ListenerRuleResource struct {
	provider Provider
}

func NewListenerRuleResource() resource.Resource {
	return &ListenerRuleResource{}
}

func (r *ListenerRuleResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *ListenerRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *ListenerRuleResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_listener_rule"
}

func (r *ListenerRuleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_listener_rule.ListenerRuleResourceSchema(ctx)
}

func (r *ListenerRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_listener_rule.ListenerRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		ListenerRuleFromTfToCreateRequest(&data),
		r.provider.ApiClient.CreateListenerRuleWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Listener Rule", err.Error())
		return
	}

	tf := ListenerRuleFromHttpToTf(res.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ListenerRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_listener_rule.ListenerRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadListenerRulesByIdResponse, error) {
		return r.provider.ApiClient.ReadListenerRulesByIdWithResponse(ctx, r.provider.SpaceID, fmt.Sprint(data.Id.ValueInt64()))
	}, http.StatusOK, &response.Diagnostics)

	tf := ListenerRuleFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ListenerRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("implement me")
}

func (r *ListenerRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_listener_rule.ListenerRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, fmt.Sprint(data.Id.ValueInt64()), r.provider.ApiClient.DeleteListenerRuleWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Listener Rule", err.Error())
		return
	}
}
