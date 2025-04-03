package managedservicebridge

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
	"terraform-provider-numspot/internal/services/managedservicebridge/datasource_managed_service_bridges"
)

var _ datasource.DataSource = &managedServiceBridgeDataSource{}

func NewManagedServiceBridgeDataSource() datasource.DataSource {
	return &managedServiceBridgeDataSource{}
}

func (d *managedServiceBridgeDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Datasource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.provider = provider
}

type managedServiceBridgeDataSource struct {
	provider *client.NumSpotSDK
}

func (d *managedServiceBridgeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_service_bridges"
}

func (d *managedServiceBridgeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_managed_service_bridges.ManagedServiceBridgesDataSourceSchema(ctx)
}

func (d *managedServiceBridgeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state, plan datasource_managed_service_bridges.ManagedServiceBridgesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, err := core.ReadManagedServiceBridges(ctx, d.provider)
	if err != nil {
		resp.Diagnostics.AddError("Error reading managed services bridges", err.Error())
		return
	}

	serviceManagedItems := serializeServiceManagedBridges(ctx, read.Items, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = serviceManagedItems.Items

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func serializeServiceManagedBridges(ctx context.Context, serviceManagedBridges []api.ManagedServicesBridge, diags *diag.Diagnostics) datasource_managed_service_bridges.ManagedServiceBridgesModel {
	serviceManageBridgeList := types.ListNull(new(datasource_managed_service_bridges.ItemsValue).Type(ctx))
	var serializeDiags diag.Diagnostics

	if len(serviceManagedBridges) > 0 {
		ll := len(serviceManagedBridges)
		itemsValue := make([]datasource_managed_service_bridges.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			itemsValue[i], serializeDiags = datasource_managed_service_bridges.NewItemsValue(datasource_managed_service_bridges.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"id": types.StringValue(serviceManagedBridges[i].Id.String()),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		serviceManageBridgeList, serializeDiags = types.ListValueFrom(ctx, new(datasource_managed_service_bridges.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}

	}

	return datasource_managed_service_bridges.ManagedServiceBridgesModel{
		Items: serviceManageBridgeList,
	}
}
