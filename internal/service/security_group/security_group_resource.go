package security_group

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-scaffolding-framework/internal/conns"
)

var _ resource.Resource = &SecurityGroupResource{}
var _ resource.ResourceWithConfigure = &SecurityGroupResource{}
var _ resource.ResourceWithImportState = &SecurityGroupResource{}

func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

type SecurityGroupResource struct {
	client *conns.ClientWithResponses
}

func (k *SecurityGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// IT SHOULD NOT BE CALLED
	var data SecurityGroupResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type SecurityGroupsMember struct {
	AccountId         basetypes.StringType `tfsdk:"account_id"`
	SecurityGroupId   basetypes.StringType `tfsdk:"security_group_id"`
	SecurityGroupName basetypes.StringType `tfsdk:"security_group_name"`
}

type RuleModel struct {
	FromPortRange basetypes.Int64Type  `tfsdk:"from_port_range"`
	ToPortRange   basetypes.Int64Type  `tfsdk:"to_port_range"`
	IpProtocol    basetypes.StringType `tfsdk:"ip_protocol"`
	ServiceIds    basetypes.StringType `tfsdk:"service_ids"`
}

type SecurityGroupResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	VirtualPrivateCloudId types.String `tfsdk:"virtual_private_cloud_id"`
	SecurityGroupName     types.String `tfsdk:"security_group_name"`
	AccountId             types.String `tfsdk:"account_id"`
	Description           types.String `tfsdk:"description"`
	InboundRules          []RuleModel  `tfsdk:"inbound_rules"`
	OutboundRules         []RuleModel  `tfsdk:"outbound_rules"`
}

func (k *SecurityGroupResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "NumSpot key pair resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The NumSpot Security Group resource computed id.",
				Computed:            true,
			},
			"virtual_private_cloud_id": schema.StringAttribute{
				MarkdownDescription: "The NumSpot Security Group Virtual Private Cloud id.",
				Required:            true,
			},
			"security_group_name": schema.StringAttribute{
				MarkdownDescription: "The NumSpot Security Group resource name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The NumSpot Security Group resource description.",
				Required:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"inbound_rules": schema.ListAttribute{
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"from_port_range": types.Int64Type,
						"to_port_range":   types.Int64Type,
						"ip_protocol":     types.StringType,
						"service_ids":     types.StringType,
					},
				},
				Computed: true,
			},
			"outbound_rules": schema.ListAttribute{
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"from_port_range": types.Int64Type,
						"to_port_range":   types.Int64Type,
						"ip_protocol":     types.StringType,
						"service_ids":     types.StringType,
					},
				},
				Computed: true,
			},
		},
	}
}

func (k *SecurityGroupResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (k *SecurityGroupResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (k *SecurityGroupResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_security_group"
}

func (k *SecurityGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data SecurityGroupResourceModel

	// Read Terraform plan data into the model
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := conns.CreateSecurityGroupJSONRequestBody{
		VirtualPrivateCloudId: data.VirtualPrivateCloudId.ValueString(),
		SecurityGroupName:     data.SecurityGroupName.ValueStringPointer(),
		Description:           data.Description.ValueStringPointer(),
	}

	createSecurityGroupResponse, err := k.client.CreateSecurityGroupWithResponse(ctx, body)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Creating Security Group (vpcId: %s)", data.VirtualPrivateCloudId.ValueString()), err.Error())
		return
	}

	numspotError := conns.HandleError(http.StatusCreated, createSecurityGroupResponse.HTTPResponse)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}

	data.Id = types.StringValue(*createSecurityGroupResponse.JSON201.Id)
	data.Description = types.StringValue(*createSecurityGroupResponse.JSON201.Description)
	data.SecurityGroupName = types.StringValue(*createSecurityGroupResponse.JSON201.SecurityGroupName)
	data.AccountId = types.StringValue(*createSecurityGroupResponse.JSON201.AccountId)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (k *SecurityGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data SecurityGroupResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	securityGroups, err := k.client.GetSecurityGroupsWithResponse(ctx)
	if err != nil {
		response.Diagnostics.AddError("Reading Security Groups", err.Error())
		return
	}

	found := false
	for _, e := range *securityGroups.JSON200.Items {
		if *e.Id == data.Id.ValueString() {
			found = true

			nData := SecurityGroupResourceModel{
				Id:                types.StringValue(*e.Id),
				SecurityGroupName: types.StringValue(*e.SecurityGroupName),
				Description:       types.StringValue(*e.Description),
				AccountId:         types.StringValue(*e.AccountId),
			}
			response.Diagnostics.Append(response.State.Set(ctx, &nData)...)
		}
	}

	if !found {
		response.State.RemoveResource(ctx)
	}
}

func (k *SecurityGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data SecurityGroupResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	res, err := k.client.DeleteSecurityGroupWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Deleting Key Pair", err.Error())
		return
	}

	numspotError := conns.HandleError(http.StatusNoContent, res.HTTPResponse)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}
}
