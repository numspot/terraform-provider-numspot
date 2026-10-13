package servercertificate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/servercertificate/resource_server_certificate"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &serverCertificateResource{}
	_ resource.ResourceWithConfigure   = &serverCertificateResource{}
	_ resource.ResourceWithImportState = &serverCertificateResource{}
)

// serverCertificateResource is the Terraform resource implementation for numspot_server_certificate.
type serverCertificateResource struct {
	services.BaseResource
}

// NewServerCertificateResource returns a new serverCertificateResource instance.
func NewServerCertificateResource() resource.Resource {
	return &serverCertificateResource{}
}

// Metadata sets the resource type name.
func (r *serverCertificateResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_server_certificate"
}

// Schema defines the schema for the resource.
func (r *serverCertificateResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_server_certificate.ServerCertificateResourceSchema(ctx)
}

// ImportState imports the resource state.
func (r *serverCertificateResource) ImportState(_ context.Context, _ resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.AddError("Import not supported", "the numspot_server_certificate resource does not support import")
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *serverCertificateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_server_certificate.ServerCertificateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numSpotServerCertificate, err := core.CreateServerCertificate(ctx, r.Provider, deserializeCreateServerCertificateRequest(&plan))
	if err != nil {
		response.Diagnostics.AddError("unable to create server_certificate", err.Error())
		return
	}

	state := serializeServerCertificate(numSpotServerCertificate)
	state.Body = utils.PreserveFromState(plan.Body)
	state.PrivateKey = utils.PreserveFromState(plan.PrivateKey)
	state.Chain = utils.PreserveFromState(plan.Chain)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *serverCertificateResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_server_certificate.ServerCertificateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotServerCertificate, err := core.ReadServerCertificate(ctx, r.Provider, state.Name.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read server_certificate", err.Error())
		return
	}

	newState := serializeServerCertificate(numSpotServerCertificate)
	newState.Body = state.Body
	newState.PrivateKey = state.PrivateKey
	newState.Chain = state.Chain

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *serverCertificateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err                      error
		state, plan              resource_server_certificate.ServerCertificateModel
		numSpotServerCertificate *api.ServerCertificate
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) || !plan.Path.Equal(state.Path) {
		numSpotServerCertificate, err = core.UpdateServerCertificate(
			ctx,
			r.Provider,
			plan.Name.ValueString(),
			deserializeUpdateServerCertificate(&state, &plan),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update server_certificate", err.Error())
			return
		}
	}

	newState := serializeServerCertificate(numSpotServerCertificate)
	newState.Body = state.Body
	newState.PrivateKey = state.PrivateKey
	newState.Chain = state.Chain

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *serverCertificateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_server_certificate.ServerCertificateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteServerCertificate(ctx, r.Provider, state.Name.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete server_certificate", err.Error())
		return
	}
}

func serializeServerCertificate(http *api.ServerCertificate) resource_server_certificate.ServerCertificateModel {
	return resource_server_certificate.ServerCertificateModel{
		ExpirationDate: types.StringPointerValue(http.ExpirationDate),
		Id:             types.StringPointerValue(http.Id),
		Name:           types.StringPointerValue(http.Name),
		Path:           types.StringPointerValue(http.Path),
		UploadDate:     types.StringPointerValue(http.UploadDate),
	}
}

func deserializeCreateServerCertificateRequest(tf *resource_server_certificate.ServerCertificateModel) *api.CreateServerCertificateJSONRequestBody {
	return &api.CreateServerCertificateJSONRequestBody{
		Body:       tf.Body.ValueString(),
		Chain:      utils.FromTfStringToStringPtr(tf.Chain),
		Name:       tf.Name.ValueString(),
		Path:       utils.FromTfStringToStringPtr(tf.Path),
		PrivateKey: tf.PrivateKey.ValueString(),
	}
}
