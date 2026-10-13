package servercertificate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/servercertificate/datasource_server_certificate"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &servercertificatesDataSource{}

// servercertificatesDataSource is the Terraform data source implementation for numspot_server_certificates.
type servercertificatesDataSource struct {
	services.BaseDataSource
}

// NewServerCertificatesDataSource returns a new servercertificatesDataSource instance.
func NewServerCertificatesDataSource() datasource.DataSource {
	return &servercertificatesDataSource{}
}

// Metadata sets the data source type name.
func (d *servercertificatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_certificates"
}

// Schema defines the schema for the data source.
func (d *servercertificatesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_server_certificate.ServerCertificateDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *servercertificatesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_server_certificate.ServerCertificateModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadServerCertificates(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadServerCertificates(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read server_certificates", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.ServerCertificate) (datasource_server_certificate.ItemsValue, diag.Diagnostics) {
	value, valueDiags := datasource_server_certificate.NewItemsValue(
		datasource_server_certificate.ItemsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"expiration_date": types.StringValue(utils.ConvertStringPtrToString(item.ExpirationDate)),
			"id":              types.StringValue(utils.ConvertStringPtrToString(item.Id)),
			"name":            types.StringValue(utils.ConvertStringPtrToString(item.Name)),
			"path":            types.StringValue(utils.ConvertStringPtrToString(item.Path)),
			"upload_date":     types.StringValue(utils.ConvertStringPtrToString(item.UploadDate)),
		},
	)
	return value, valueDiags
}

func deserializeReadServerCertificates(
	ctx context.Context,
	tf *datasource_server_certificate.ServerCertificateModel,
	diags *diag.Diagnostics,
) *api.ReadServerCertificatesParams {
	return &api.ReadServerCertificatesParams{
		Paths: utils.ConvertTfListToArrayOfString(ctx, tf.Paths, diags),
	}
}
