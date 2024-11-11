package vpcpeering

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type VpcPeeringsDataSourceModel struct {
	Items               []VpcPeeringDatasourceItemModel `tfsdk:"items"`
	AccepterVpcIpRanges types.List                      `tfsdk:"accepter_vpc_ip_ranges"`
	AccepterVpcVpcIds   types.List                      `tfsdk:"accepter_vpc_vpc_ids"`
	ExpirationDates     types.List                      `tfsdk:"expiration_dates"`
	Ids                 types.List                      `tfsdk:"ids"`
	SourceVpcIpRanges   types.List                      `tfsdk:"source_vpc_ip_ranges"`
	SourceVpcVpcIds     types.List                      `tfsdk:"source_vpc_vpc_ids"`
	StateMessages       types.List                      `tfsdk:"state_messages"`
	StateNames          types.List                      `tfsdk:"state_names"`
	TagKeys             types.List                      `tfsdk:"tag_keys"`
	TagValues           types.List                      `tfsdk:"tag_values"`
	Tags                types.List                      `tfsdk:"tags"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &vpcPeeringsDataSource{}
)

func (d *vpcPeeringsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVpcPeeringsDataSource() datasource.DataSource {
	return &vpcPeeringsDataSource{}
}

type vpcPeeringsDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *vpcPeeringsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_peerings"
}

// Schema defines the schema for the data source.
func (d *vpcPeeringsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = VpcPeeringDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *vpcPeeringsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VpcPeeringsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeVpcPeeringDatasource(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	vpcPeerings, err := core.ReadVPCPeerings(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("failed to read VPC peerings", err.Error())
		return
	}

	objectItems := serializeVpcPeeringDatasource(ctx, vpcPeerings, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func serializeVpcPeeringDatasource(ctx context.Context, vpcPeerings *[]numspot.VpcPeering, diags *diag.Diagnostics) []VpcPeeringDatasourceItemModel {
	serializeVPCPeeringDatasourceItem := func(
		ctx context.Context,
		http *numspot.VpcPeering,
		diags *diag.Diagnostics,
	) *VpcPeeringDatasourceItemModel {
		var (
			tagsList         types.List
			accepterVpc      AccepterVpcValue
			sourceVpc        SourceVpcValue
			state            StateValue
			expirationDateTf types.String
		)

		if http.Tags != nil {
			tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
			if diags.HasError() {
				return nil
			}
		}

		if http.ExpirationDate != nil {
			expirationDate := *http.ExpirationDate
			expirationDateTf = types.StringValue(expirationDate.Format(time.RFC3339))
		}

		if http.AccepterVpc != nil {
			var diagnostics diag.Diagnostics
			accepterVpc, diagnostics = NewAccepterVpcValue(AccepterVpcValue{}.AttributeTypes(ctx),
				map[string]attr.Value{
					"ip_range": types.StringPointerValue(http.AccepterVpc.IpRange),
					"vpc_id":   types.StringPointerValue(http.AccepterVpc.VpcId),
				})
			diags.Append(diagnostics...)
		}

		if http.SourceVpc != nil {
			var diagnostics diag.Diagnostics
			sourceVpc, diagnostics = NewSourceVpcValue(SourceVpcValue{}.AttributeTypes(ctx),
				map[string]attr.Value{
					"ip_range": types.StringPointerValue(http.SourceVpc.IpRange),
					"vpc_id":   types.StringPointerValue(http.SourceVpc.VpcId),
				})
			diags.Append(diagnostics...)
		}

		if http.State != nil {
			var diagnostics diag.Diagnostics
			state, diagnostics = NewStateValue(StateValue{}.AttributeTypes(ctx),
				map[string]attr.Value{
					"message": types.StringPointerValue(http.State.Message),
					"name":    types.StringPointerValue(http.State.Name),
				})
			diags.Append(diagnostics...)
		}

		return &VpcPeeringDatasourceItemModel{
			Id:             types.StringPointerValue(http.Id),
			Tags:           tagsList,
			AccepterVpc:    accepterVpc,
			ExpirationDate: expirationDateTf,
			SourceVpc:      sourceVpc,
			State:          state,
		}
	}
	return utils.FromHttpGenericListToTfList(ctx, vpcPeerings, serializeVPCPeeringDatasourceItem, diags)
}

func deserializeVpcPeeringDatasource(ctx context.Context, tf VpcPeeringsDataSourceModel, diags *diag.Diagnostics) numspot.ReadVpcPeeringsParams {
	expirationDates := utils.TfStringListToTimeList(ctx, tf.ExpirationDates, "2020-06-30T00:00:00.000Z", diags)

	return numspot.ReadVpcPeeringsParams{
		ExpirationDates:     &expirationDates,
		StateMessages:       utils.TfStringListToStringPtrList(ctx, tf.StateMessages, diags),
		StateNames:          utils.TfStringListToStringPtrList(ctx, tf.StateNames, diags),
		AccepterVpcIpRanges: utils.TfStringListToStringPtrList(ctx, tf.AccepterVpcIpRanges, diags),
		AccepterVpcVpcIds:   utils.TfStringListToStringPtrList(ctx, tf.AccepterVpcVpcIds, diags),
		Ids:                 utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		SourceVpcIpRanges:   utils.TfStringListToStringPtrList(ctx, tf.SourceVpcIpRanges, diags),
		SourceVpcVpcIds:     utils.TfStringListToStringPtrList(ctx, tf.SourceVpcVpcIds, diags),
		TagKeys:             utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:           utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:                utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
	}
}
