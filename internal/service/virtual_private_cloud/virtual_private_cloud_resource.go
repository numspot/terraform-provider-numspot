package virtual_private_cloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-scaffolding-framework/internal/conns"
)

var _ resource.Resource = &VirtualPrivateCloudResource{}
var _ resource.ResourceWithConfigure = &VirtualPrivateCloudResource{}
var _ resource.ResourceWithImportState = &VirtualPrivateCloudResource{}

func NewVirtualPrivateCloudResource() resource.Resource {
	return &VirtualPrivateCloudResource{}
}

type VirtualPrivateCloudResource struct {
	client *conns.ClientWithResponses
}

func (k *VirtualPrivateCloudResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// IT SHOULD NOT BE CALLED
	var data VirtualPrivateCloudResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type VirtualPrivateCloudResourceModel struct {
	Id               types.String `tfsdk:"id"`
	IpRange          types.String `tfsdk:"ip_range"`
	State            types.String `tfsdk:"state"`
	DhcpOptionsSetId types.String `tfsdk:"dhcp_options_set_id"`
	Tenancy          types.String `tfsdk:"tenancy"`
}

func (k *VirtualPrivateCloudResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "NumSpot key pair resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The NumSpot VPC resource computed id. It is equal to the 'virtual_private_cloud_id' field.",
				Computed:            true,
			},
			"ip_range": schema.StringAttribute{
				MarkdownDescription: "The NumSpot VPC resource ip range",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The NumSpot VPC resource computed state.",
				Computed:            true,
			},
			"dhcp_options_set_id": schema.StringAttribute{
				MarkdownDescription: "The NumSpot VPC resource computed DHCP Options Set id.",
				Computed:            true,
			},
			"tenancy": schema.StringAttribute{
				MarkdownDescription: "The NumSpot VPC resource DHCP tenancy.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (k *VirtualPrivateCloudResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (k *VirtualPrivateCloudResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*conns.ClientWithResponses)

	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	k.client = client
}

func (k *VirtualPrivateCloudResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpc"
}

func (k *VirtualPrivateCloudResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data VirtualPrivateCloudResourceModel

	// Read Terraform plan data into the model
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	body := conns.CreateVPCJSONRequestBody{
		IpRange: data.IpRange.ValueString(),
		Tenancy: data.Tenancy.ValueStringPointer(),
	}

	createVPCResponse, err := k.client.CreateVPCWithResponse(ctx, body)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Creating VPC (%s)", data.IpRange.ValueString()), err.Error())
		return
	}

	numspotError := conns.HandleError(http.StatusCreated, createVPCResponse.HTTPResponse)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}

	data.Id = types.StringValue(*createVPCResponse.JSON201.Id)
	data.IpRange = types.StringValue(*createVPCResponse.JSON201.IpRange)
	data.Tenancy = types.StringValue(*createVPCResponse.JSON201.Tenancy)
	data.DhcpOptionsSetId = types.StringValue(*createVPCResponse.JSON201.DhcpOptionsSetId)
	data.State = types.StringValue(*createVPCResponse.JSON201.State)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (k *VirtualPrivateCloudResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data VirtualPrivateCloudResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	keyPairs, err := k.client.GetVPCWithResponse(ctx)
	if err != nil {
		response.Diagnostics.AddError("Reading Key Pairs", err.Error())
		return
	}

	found := false
	for _, e := range *keyPairs.JSON200.Items {
		if *e.Id == data.Id.ValueString() {
			found = true

			nData := VirtualPrivateCloudResourceModel{
				Id:               types.StringValue(*e.Id),
				IpRange:          types.StringValue(*e.IpRange),
				Tenancy:          types.StringValue(*e.Tenancy),
				State:            types.StringValue(*e.State),
				DhcpOptionsSetId: types.StringValue(*e.DhcpOptionsSetId),
			}
			response.Diagnostics.Append(response.State.Set(ctx, &nData)...)
		}
	}

	if !found {
		response.State.RemoveResource(ctx)
	}
}

func (k *VirtualPrivateCloudResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VirtualPrivateCloudResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	res, err := k.client.DeleteVPCWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Deleting VPC", err.Error())
		return
	}

	numspotError := conns.HandleError(http.StatusNoContent, res.HTTPResponse)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}
}
