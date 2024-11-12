package virtualgateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type VirtualGatewaysDataSourceModel struct {
	Items           []VirtualGatewayModelItemDataSource `tfsdk:"items"`
	ConnectionTypes types.List                          `tfsdk:"connection_types"`
	LinkStates      types.List                          `tfsdk:"link_states"`
	States          types.List                          `tfsdk:"states"`
	TagKeys         types.List                          `tfsdk:"tag_keys"`
	TagValues       types.List                          `tfsdk:"tag_values"`
	Tags            types.List                          `tfsdk:"tags"`
	LinkVpcIds      types.List                          `tfsdk:"link_vpc_ids"`
	IDs             types.List                          `tfsdk:"ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &virtualGatewaysDataSource{}
)

func (d *virtualGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVirtualGatewaysDataSource() datasource.DataSource {
	return &virtualGatewaysDataSource{}
}

type virtualGatewaysDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *virtualGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_gateways"
}

// Schema defines the schema for the data source.
func (d *virtualGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = VirtualGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *virtualGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VirtualGatewaysDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeVirtualGatewayParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numspotVirtualGateway, err := core.ReadVirtualGatewaysWithParams(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read virtual gateways", err.Error())
		return
	}

	objectItems := serializeVirtualGatewayDatasource(ctx, numspotVirtualGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func serializeVirtualGatewayDatasource(ctx context.Context, virtualGateways *[]numspot.VirtualGateway, diags *diag.Diagnostics) []VirtualGatewayModelItemDataSource {
	return utils.FromHttpGenericListToTfList(ctx, virtualGateways, func(ctx context.Context, http *numspot.VirtualGateway, diags *diag.Diagnostics) *VirtualGatewayModelItemDataSource {
		var tagsTf, vpcToVirtualGatewayLinksTf types.List

		if http.Tags != nil {
			tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
			if diags.HasError() {
				return nil
			}
		}

		if http.VpcToVirtualGatewayLinks != nil {
			vpcToVirtualGatewayLinksTf = utils.GenericListToTfListValue(ctx, serializeVpcToVirtualGatewayLinks, *http.VpcToVirtualGatewayLinks, diags)
			if diags.HasError() {
				return nil
			}
		}

		return &VirtualGatewayModelItemDataSource{
			ConnectionType:           types.StringPointerValue(http.ConnectionType),
			Id:                       types.StringPointerValue(http.Id),
			VpcToVirtualGatewayLinks: vpcToVirtualGatewayLinksTf,
			State:                    types.StringPointerValue(http.State),
			Tags:                     tagsTf,
		}
	}, diags)
}

func deserializeVirtualGatewayParams(ctx context.Context, tf VirtualGatewaysDataSourceModel, diags *diag.Diagnostics) numspot.ReadVirtualGatewaysParams {
	return numspot.ReadVirtualGatewaysParams{
		States:          utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		TagKeys:         utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:       utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:            utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:             utils.TfStringListToStringPtrList(ctx, tf.IDs, diags),
		ConnectionTypes: utils.TfStringListToStringPtrList(ctx, tf.ConnectionTypes, diags),
		LinkStates:      utils.TfStringListToStringPtrList(ctx, tf.LinkStates, diags),
		LinkVpcIds:      utils.TfStringListToStringPtrList(ctx, tf.LinkVpcIds, diags),
	}
}
