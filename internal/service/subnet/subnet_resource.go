package subnet

import (
	"context"
	"fmt"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns"
	api_client "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api_client"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &SubnetResource{}
	_ resource.ResourceWithConfigure   = &SubnetResource{}
	_ resource.ResourceWithImportState = &SubnetResource{}
)

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

type SubnetResource struct {
	client *api_client.ClientWithResponses
}

type SubnetResourceModel struct {
	Id                  types.String `tfsdk:"id"`
	IpRange             types.String `tfsdk:"ip_range"`
	State               types.String `tfsdk:"state"`
	VpcId               types.String `tfsdk:"virtual_private_cloud_id"`
	AvailabilityZone    types.String `tfsdk:"availability_zone"`
	AvailableIpsCount   types.Int64  `tfsdk:"available_ips_count"`
	MapPublicIpOnLaunch types.Bool   `tfsdk:"map_public_ip_on_launch"`
}

func (k *SubnetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data SubnetResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Get(ctx, &data)...)
}

func (k *SubnetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "NumSpot subnet resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The NumSpot subnet resource computed id.",
				Computed:            true,
			},
			"ip_range": schema.StringAttribute{
				MarkdownDescription: "The list of network prefixes used by the NumSpot subnet.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the NumSpot subnet.",
				Computed:            true,
				Optional:            true,
			},
			"virtual_private_cloud_id": schema.StringAttribute{
				MarkdownDescription: "The id of the parent NumSpot VPC in which the Subnet is.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The name of the region in which the NunSpot subnet is located.",
				Computed:            true,
				Optional:            true,
			},
			"available_ips_count": schema.Int64Attribute{
				MarkdownDescription: "The number of available IPs in the Subnets.",
				Computed:            true,
				Optional:            true,
			},
			"map_public_ip_on_launch": schema.BoolAttribute{
				MarkdownDescription: "If true, a public IP is assigned to the network interface cards (NICs) created in the NumSpot Subnet.",
				Computed:            true,
				Optional:            true,
			},
		},
	}
}

func (k *SubnetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (k *SubnetResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	client, ok := request.ProviderData.(*api_client.ClientWithResponses)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type", fmt.Sprintf("%T", request.ProviderData),
		)
		return
	}
	k.client = client
}

func (k *SubnetResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_subnet"
}

func (k *SubnetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data SubnetResourceModel

	// Read Terraform plan data into the model
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := api_client.CreateSubnetJSONRequestBody{
		IpRange:               data.IpRange.ValueString(),
		VirtualPrivateCloudId: data.VpcId.ValueString(),
		AvailabilityZone:      data.AvailabilityZone.ValueStringPointer(),
	}

	createSubnetResponse, err := k.client.CreateSubnetWithResponse(ctx, body)
	if err != nil {
		response.Diagnostics.AddError("Creating Subnet", err.Error())
		return
	}

	numspotError := conns.HandleErrorBis(http.StatusCreated, createSubnetResponse.HTTPResponse.StatusCode, createSubnetResponse.Body)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}

	// Read SubnetResponse into the model
	data.importResponse(createSubnetResponse)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (k *SubnetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data SubnetResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	res, err := k.client.GetSubnetsWithResponse(ctx)
	if err != nil {
		response.Diagnostics.AddError("Reading Subnet", err.Error())
		return
	}

	numspotError := conns.HandleErrorBis(http.StatusNoContent, res.HTTPResponse.StatusCode, res.Body)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}

	found := false
	for _, e := range *res.JSON200.Items {
		if *e.Id == data.Id.ValueString() {
			found = true

			nData := SubnetResourceModel{
				Id:                  types.StringValue(*e.Id),
				IpRange:             types.StringValue(e.IpRange),
				State:               types.StringValue(*e.State),
				VpcId:               types.StringValue(*e.VirtualPrivateCloudId),
				AvailabilityZone:    types.StringValue(*e.AvailabilityZone),
				AvailableIpsCount:   types.Int64Value(int64(*e.AvailableIpsCount)),
				MapPublicIpOnLaunch: types.BoolValue(*e.MapPublicIpOnLaunch),
			}
			response.Diagnostics.Append(response.State.Set(ctx, &nData)...)
		}
	}
	if !found {
		response.State.RemoveResource(ctx)
	}
}

func (k *SubnetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data SubnetResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := k.client.DeleteSubnetWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Deleting Subnet", err.Error())
		return
	}

	numspotError := conns.HandleErrorBis(http.StatusNoContent, res.HTTPResponse.StatusCode, res.Body)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}
}

func (data *SubnetResourceModel) importResponse(r *api_client.CreateSubnetResponse) {
	data.Id = types.StringValue(*r.JSON201.Id)
	data.IpRange = types.StringValue(r.JSON201.IpRange)
	data.State = types.StringValue(*r.JSON201.State)
	data.VpcId = types.StringValue(*r.JSON201.VirtualPrivateCloudId)
	data.AvailabilityZone = types.StringValue(*r.JSON201.AvailabilityZone)
	data.AvailableIpsCount = types.Int64Value(int64(*r.JSON201.AvailableIpsCount))
	data.MapPublicIpOnLaunch = types.BoolValue(*r.JSON201.MapPublicIpOnLaunch)
}
