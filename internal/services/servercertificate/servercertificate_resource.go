package servercertificate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/servercertificate/resource_server_certificate"
)

var _ resource.Resource = &serverCertificateResource{}

type serverCertificateResource struct {
	provider *client.NumSpotSDK
}

func NewServerCertificateResource() resource.Resource {
	return &serverCertificateResource{}
}

func (r *serverCertificateResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *serverCertificateResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *serverCertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_certificate"
}

func (r *serverCertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_server_certificate.ServerCertificateResourceSchema(ctx)
}

func (r *serverCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_server_certificate.ServerCertificateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := deserializeServerCertificate(plan)

	numSpot, err := core.CreateServerCertificate(ctx, r.provider, body)
	if err != nil {
		resp.Diagnostics.AddError("unable to create server certificate", err.Error())
		return
	}

	data := serializeServerCertificate(numSpot, body.Body, body.PrivateKey, body.Chain)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan resource_server_certificate.ServerCertificateModel

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.Body.ValueString()
	chain := plan.Chain.ValueStringPointer()
	key := plan.PrivateKey.ValueString()
	serverCertificateName := plan.Name.ValueString()

	serverCertificate, err := core.ReadServerCertificate(ctx, r.provider, serverCertificateName)
	if err != nil {
		resp.Diagnostics.AddError("unable to read server certificate", err.Error())
		return
	}

	newPlan := serializeServerCertificate(serverCertificate, body, key, chain)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newPlan)...)
}

func (r *serverCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		err               error
		serverCertificate *api.ServerCertificate
		state, plan       resource_server_certificate.ServerCertificateModel
	)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.Body.ValueString()
	chain := plan.Chain.ValueStringPointer()
	key := plan.PrivateKey.ValueString()
	needUpdate := false
	updateBody := api.UpdateServerCertificateJSONRequestBody{}

	if plan.Name != state.Name {
		needUpdate = true
		updateBody.NewName = plan.Name.ValueStringPointer()
	}

	if plan.Path != state.Path {
		needUpdate = true
		updateBody.NewPath = plan.Path.ValueStringPointer()
	}

	if needUpdate {
		serverCertificate, err = core.UpdateServerCertificate(ctx, r.provider, plan.Name.ValueString(), updateBody)
		if err != nil {
			resp.Diagnostics.AddError("unable to update server certificate", err.Error())
			return
		}

		newState := serializeServerCertificate(serverCertificate, body, key, chain)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	}
}

func (r *serverCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan resource_server_certificate.ServerCertificateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverCertificateName := plan.Name.ValueString()

	if err := core.DeleteServerCertificate(ctx, r.provider, serverCertificateName); err != nil {
		resp.Diagnostics.AddError("unable to delete server certificate", err.Error())
		return
	}
}

func deserializeServerCertificate(tf resource_server_certificate.ServerCertificateModel) api.CreateServerCertificateJSONRequestBody {
	var chain, path *string = nil, nil

	if tf.Chain.IsNull() {
		chain = tf.Chain.ValueStringPointer()
	}

	if tf.Path.IsNull() {
		path = tf.Path.ValueStringPointer()
	}

	return api.CreateServerCertificateJSONRequestBody{
		Body:       tf.Body.ValueString(),
		Chain:      chain,
		Name:       tf.Name.ValueString(),
		Path:       path,
		PrivateKey: tf.PrivateKey.ValueString(),
	}
}

func serializeServerCertificate(http *api.ServerCertificate, body, key string, chain *string) resource_server_certificate.ServerCertificateModel {
	return resource_server_certificate.ServerCertificateModel{
		Body:           types.StringValue(body),
		Chain:          types.StringPointerValue(chain),
		ExpirationDate: types.StringPointerValue(http.ExpirationDate),
		Id:             types.StringPointerValue(http.Id),
		Name:           types.StringPointerValue(http.Name),
		Path:           types.StringPointerValue(http.Path),
		PrivateKey:     types.StringValue(key),
		UploadDate:     types.StringPointerValue(http.UploadDate),
	}
}
