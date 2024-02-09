package provider

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_listener_rule"
)

var (
	_ resource.Resource                = &ListenerRuleResource{}
	_ resource.ResourceWithConfigure   = &ListenerRuleResource{}
	_ resource.ResourceWithImportState = &ListenerRuleResource{}
)

type ListenerRuleResource struct {
	client *api.ClientWithResponses
}

func NewListenerRuleResource() resource.Resource {
	return &ListenerRuleResource{}
}

func (r *ListenerRuleResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*api.ClientWithResponses)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.client = client
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

	res := utils.ExecuteRequest(func() (*api.CreateListenerRuleResponse, error) {
		body := ListenerRuleFromTfToCreateRequest(&data)
		return r.client.CreateListenerRuleWithResponse(ctx, body)
	}, http.StatusOK, &response.Diagnostics)

	tf := ListenerRuleFromHttpToTf(res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *ListenerRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_listener_rule.ListenerRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadListenerRulesByIdResponse, error) {
		return r.client.ReadListenerRulesByIdWithResponse(ctx, fmt.Sprint(data.Id.ValueInt64()))
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

	utils.ExecuteRequest(func() (*api.DeleteListenerRuleResponse, error) {
		return r.client.DeleteListenerRuleWithResponse(ctx, fmt.Sprint(data.Id.ValueInt64()))
	}, http.StatusOK, &response.Diagnostics)
}
