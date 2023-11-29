package dhcp_options_set

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns"
	api_client "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api_client"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DhcpOptionsSetResource{}
var _ resource.ResourceWithConfigure = &DhcpOptionsSetResource{}
var _ resource.ResourceWithImportState = &DhcpOptionsSetResource{}

func NewDhcpOptionsSetResource() resource.Resource {
	return &DhcpOptionsSetResource{}
}

type DhcpOptionsSetResource struct {
	client *api_client.ClientWithResponses
}

type DhcpOptionsSetResourceModel struct {
	Id                types.String `tfsdk:"id"`
	DomainName        types.String `tfsdk:"domain_name"`
	DomainNameServers types.List   `tfsdk:"domain_name_servers"`
	LogServers        types.List   `tfsdk:"log_servers"`
	NtpServers        types.List   `tfsdk:"ntp_servers"`
}

func (k *DhcpOptionsSetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// IT SHOULD NOT BE CALLED
	var data DhcpOptionsSetResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (k *DhcpOptionsSetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "NumSpot key pair resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The NumSpot VPC resource computed id. It is equal to the 'virtual_private_cloud_id' field.",
				Computed:            true,
			},
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The NumSpot VPC resource ip range",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_name_servers": schema.ListAttribute{
				MarkdownDescription: "The NumSpot VPC resource computed state.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"log_servers": schema.ListAttribute{
				MarkdownDescription: "The NumSpot VPC resource computed state.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"ntp_servers": schema.ListAttribute{
				MarkdownDescription: "The NumSpot VPC resource computed state.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (k *DhcpOptionsSetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	pr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, pr, request, response)
}

func (k *DhcpOptionsSetResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*api_client.ClientWithResponses)

	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	k.client = client
}

func (k *DhcpOptionsSetResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_dhcp_options_set"
}

func (k *DhcpOptionsSetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data DhcpOptionsSetResourceModel

	// Read Terraform plan data into the model
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	body := api_client.CreateDhcpOptionsSetJSONRequestBody{
		DomainName: data.DomainName.ValueStringPointer(),
	}

	// Domain Name Servers
	domainNameServersElements := make([]string, 0, len(data.DomainNameServers.Elements()))
	diags := data.DomainNameServers.ElementsAs(ctx, &domainNameServersElements, true)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	body.DomainNameServers = &domainNameServersElements

	// Log Servers
	logServersElements := make([]string, 0, len(data.LogServers.Elements()))
	diags = data.LogServers.ElementsAs(ctx, &logServersElements, true)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	body.LogServers = &logServersElements

	// Ntp Servers
	ntpServersElements := make([]string, 0, len(data.NtpServers.Elements()))
	diags = data.NtpServers.ElementsAs(ctx, &ntpServersElements, true)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	body.NtpServers = &ntpServersElements

	createDhcpOptionsSetResponse, err := k.client.CreateDhcpOptionsSet(ctx, body)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Creating Dhcp Options Set (%s)", data.Id.ValueString()), err.Error())
		return
	}

	numspotError := conns.HandleError(http.StatusCreated, createDhcpOptionsSetResponse)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}

	createDhcpOptionsSet, err := api_client.ParseCreateDhcpOptionsSetResponse(createDhcpOptionsSetResponse)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Parsing Dhcp Options Set (%s)", data.Id.ValueString()), err.Error())
		return
	}

	data.Id = types.StringValue(*createDhcpOptionsSet.JSON201.Id)
	data.DomainName = types.StringValue(*createDhcpOptionsSet.JSON201.DomainName)

	// Domain Name Servers
	domainNameServers, diag := types.ListValueFrom(ctx, types.StringType, createDhcpOptionsSet.JSON201.DomainNameServers)
	response.Diagnostics.Append(diag...)
	if response.Diagnostics.HasError() {
		return
	}
	data.DomainNameServers = domainNameServers

	// Log Servers
	logServers, diag := types.ListValueFrom(ctx, types.StringType, createDhcpOptionsSet.JSON201.LogServers)
	response.Diagnostics.Append(diag...)
	if response.Diagnostics.HasError() {
		return
	}
	data.LogServers = logServers

	// Ntp Servers
	ntpServers, diag := types.ListValueFrom(ctx, types.StringType, createDhcpOptionsSet.JSON201.NtpServers)
	response.Diagnostics.Append(diag...)
	if response.Diagnostics.HasError() {
		return
	}
	data.NtpServers = ntpServers

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (k *DhcpOptionsSetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data DhcpOptionsSetResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	dhcpOptionsSet, err := k.client.GetDhcpOptionsSetWithResponse(ctx)
	if err != nil {
		response.Diagnostics.AddError("Reading Key Pairs", err.Error())
		return
	}

	found := false
	for _, e := range *dhcpOptionsSet.JSON200.Items {
		// TODO: Complete compare Objects
		if *e.Id == data.Id.ValueString() {
			found = true

			nData := DhcpOptionsSetResourceModel{
				Id:         types.StringValue(*e.Id),
				DomainName: types.StringValue(*e.DomainName),
			}

			// Domain Name Servers
			domainNameServers, diag := types.ListValueFrom(ctx, types.StringType, e.DomainNameServers)
			response.Diagnostics.Append(diag...)
			if response.Diagnostics.HasError() {
				return
			}
			nData.DomainNameServers = domainNameServers

			// Log Servers
			logServers, diag := types.ListValueFrom(ctx, types.StringType, e.DomainNameServers)
			response.Diagnostics.Append(diag...)
			if response.Diagnostics.HasError() {
				return
			}
			nData.LogServers = logServers

			// Ntp Servers
			ntpServers, diag := types.ListValueFrom(ctx, types.StringType, e.NtpServers)
			response.Diagnostics.Append(diag...)
			if response.Diagnostics.HasError() {
				return
			}
			nData.NtpServers = ntpServers

			response.Diagnostics.Append(response.State.Set(ctx, &nData)...)
		}
	}

	if !found {
		response.State.RemoveResource(ctx)
	}
}

func (k *DhcpOptionsSetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data DhcpOptionsSetResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	res, err := k.client.DeleteDhcpOptionsSetWithResponse(ctx, data.Id.ValueString())
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
