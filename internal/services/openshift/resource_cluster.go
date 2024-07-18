package openshift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
)

var (
	_ resource.Resource                = &ClusterResource{}
	_ resource.ResourceWithConfigure   = &ClusterResource{}
	_ resource.ResourceWithImportState = &ClusterResource{}
)

type ClusterResource struct {
	provider services.IProvider
}

func (r *ClusterResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *ClusterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *ClusterResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_ocp_cluster"
}

func (r *ClusterResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = ClusterResourceSchema(ctx)
}

func (r *ClusterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// TODO implement me
	panic("implement me")
}

func (r *ClusterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// TODO implement me
	panic("implement me")
}

func (r *ClusterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// TODO implement me
	panic("implement me")
}

func (r *ClusterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// TODO implement me
	panic("implement me")
}
