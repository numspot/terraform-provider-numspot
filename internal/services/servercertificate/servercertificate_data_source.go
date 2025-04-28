package servercertificate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/servercertificate/datasource_server_certificate"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &servercertificateDataSource{}

func (d *servercertificateDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	d.provider = services.ConfigureProviderDatasource(request, response)
}

func NewServerCertificateDataSource() datasource.DataSource {
	return &servercertificateDataSource{}
}

type servercertificateDataSource struct {
	provider *client.NumSpotSDK
}

func (d *servercertificateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_certificate"
}

func (d *servercertificateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_server_certificate.ServerCertificateDataSourceSchema(ctx)
}

func (d *servercertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state, plan datasource_server_certificate.ServerCertificateModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverCertificateParams := deserializeReadServerCertificate(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	read, err := core.ReadServerCertificates(ctx, d.provider, &serverCertificateParams)
	if err != nil {
		resp.Diagnostics.AddError("unable to read server certificate", err.Error())
		return
	}

	serverCertificateItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *read, mappingItemsValue)
	if serializeDiags.HasError() {
		resp.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, serverCertificateItems, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func deserializeReadServerCertificate(ctx context.Context, tf datasource_server_certificate.ServerCertificateModel, diags *diag.Diagnostics) api.ReadServerCertificatesParams {
	return api.ReadServerCertificatesParams{
		Paths: utils.ConvertTfListToArrayOfString(ctx, tf.Paths, diags),
	}
}

func mappingItemsValue(ctx context.Context, serverCertificate api.ServerCertificate) (datasource_server_certificate.ItemsValue, diag.Diagnostics) {
	return datasource_server_certificate.NewItemsValue(datasource_server_certificate.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"name":            types.StringValue(utils.ConvertStringPtrToString(serverCertificate.Name)),
		"expiration_date": types.StringValue(utils.ConvertStringPtrToString(serverCertificate.ExpirationDate)),
		"id":              types.StringValue(utils.ConvertStringPtrToString(serverCertificate.Id)),
		"path":            types.StringValue(utils.ConvertStringPtrToString(serverCertificate.Path)),
		"upload_date":     types.StringValue(utils.ConvertStringPtrToString(serverCertificate.UploadDate)),
	})
}
