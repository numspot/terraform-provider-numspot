package servercertificate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/servercertificate/datasource_server_certificate"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &servercertificateDataSource{}

func (d *servercertificateDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.provider = provider
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

	serverCertificateItems := serializeServerCertificates(ctx, read, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = serverCertificateItems.Items

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func deserializeReadServerCertificate(ctx context.Context, tf datasource_server_certificate.ServerCertificateModel, diags *diag.Diagnostics) api.ReadServerCertificatesParams {
	return api.ReadServerCertificatesParams{
		Paths: utils.ConvertTfListToArrayOfString(ctx, tf.Paths, diags),
	}
}

func serializeServerCertificates(ctx context.Context, serverCertificate *[]api.ServerCertificate, diags *diag.Diagnostics) datasource_server_certificate.ServerCertificateModel {
	var serverCertificateList types.List
	var serializeDiags diag.Diagnostics

	if len(*serverCertificate) != 0 {
		ll := len(*serverCertificate)
		itemsValue := make([]datasource_server_certificate.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			itemsValue[i], serializeDiags = datasource_server_certificate.NewItemsValue(datasource_server_certificate.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"name":            types.StringValue(utils.ConvertStringPtrToString((*serverCertificate)[i].Name)),
				"expiration_date": types.StringValue(utils.ConvertStringPtrToString((*serverCertificate)[i].ExpirationDate)),
				"id":              types.StringValue(utils.ConvertStringPtrToString((*serverCertificate)[i].Id)),
				"path":            types.StringValue(utils.ConvertStringPtrToString((*serverCertificate)[i].Path)),
				"upload_date":     types.StringValue(utils.ConvertStringPtrToString((*serverCertificate)[i].UploadDate)),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		serverCertificateList, serializeDiags = types.ListValueFrom(ctx, new(datasource_server_certificate.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		serverCertificateList = types.ListNull(new(datasource_server_certificate.ItemsValue).Type(ctx))
	}

	return datasource_server_certificate.ServerCertificateModel{
		Items: serverCertificateList,
	}
}
