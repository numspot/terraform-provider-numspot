package services

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-numspot/internal/client"
)

// BaseResource is the Terraform resource implementation.
type BaseResource struct {
	Provider *client.NumSpotSDK
}

// Configure configures the provider data.
func (b *BaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	b.Provider = ConfigureProviderResource(req, resp)
}

// ImportState imports the state.
func (b *BaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
