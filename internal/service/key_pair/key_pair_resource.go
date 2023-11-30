package key_pair

import (
	"context"
	"fmt"
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

var _ resource.Resource = &KeyPairResource{}
var _ resource.ResourceWithConfigure = &KeyPairResource{}
var _ resource.ResourceWithImportState = &KeyPairResource{}

func NewKeyPairResource() resource.Resource {
	return &KeyPairResource{}
}

type KeyPairResource struct {
	client *api_client.ClientWithResponses
}

func (k *KeyPairResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// IT SHOULD NOT BE CALLED
	var data KeyPairResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type KeyPairResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PublicKey   types.String `tfsdk:"public_key"`
	PrivateKey  types.String `tfsdk:"private_key"`
	Fingerprint types.String `tfsdk:"fingerprint"`
}

func (k *KeyPairResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "NumSpot key pair resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The NumSpot key pair resource computed id.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The NumSpot key pair resource name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "The NumSpot key pair resource public key",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "The NumSpot key pair resource computed private key.",
				Sensitive:           true,
				Computed:            true,
				Optional:            true,
			},
			"fingerprint": schema.StringAttribute{
				MarkdownDescription: "The NumSpot key pair resource computed fingerprint.",
				Computed:            true,
			},
		},
	}
}

func (k *KeyPairResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), request, response)
}

func (k *KeyPairResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (k *KeyPairResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_key_pair"
}

func (k *KeyPairResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data KeyPairResourceModel

	// Read Terraform plan data into the model
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	isNotImport := data.PublicKey.IsNull()
	if isNotImport {
		body := api_client.CreateKeyPairJSONRequestBody{
			Name: data.Name.ValueString(),
		}

		res, err := k.client.CreateKeyPairWithResponse(ctx, body)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("Creating Key Pair (%s)", data.Name.ValueString()), err.Error())
			return
		}

		numspotError := conns.HandleError(http.StatusCreated, res.HTTPResponse.StatusCode, res.Body)
		if numspotError != nil {
			response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
			return
		}

		data.Id = types.StringValue(*res.JSON201.Name)
		data.Name = types.StringValue(*res.JSON201.Name)
		data.PrivateKey = types.StringValue(*res.JSON201.PrivateKey)
		data.Fingerprint = types.StringValue(*res.JSON201.Fingerprint)

		// Save data into Terraform state
		response.Diagnostics.Append(response.State.Set(ctx, &data)...)
	} else {
		body := api_client.ImportKeyPairJSONRequestBody{
			Name:      data.Name.ValueString(),
			PublicKey: data.PublicKey.ValueString(),
		}

		res, err := k.client.ImportKeyPairWithResponse(ctx, body)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("Importing Key Pair (%s)", data.Name.ValueString()), err.Error())
			return
		}

		numspotError := conns.HandleError(http.StatusCreated, res.HTTPResponse.StatusCode, res.Body)
		if numspotError != nil {
			response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
			return
		}

		data.Id = types.StringValue(*res.JSON201.Name)
		data.Name = types.StringValue(*res.JSON201.Name)
		data.Fingerprint = types.StringValue(*res.JSON201.Fingerprint)
		data.PrivateKey = types.StringNull()

		// Save data into Terraform state
		response.Diagnostics.Append(response.State.Set(ctx, &data)...)
	}
}

func (k *KeyPairResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data KeyPairResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	res, err := k.client.GetKeyPairsWithResponse(ctx)
	if err != nil {
		response.Diagnostics.AddError("Reading Key Pairs", err.Error())
		return
	}

	numspotError := conns.HandleError(http.StatusOK, res.HTTPResponse.StatusCode, res.Body)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}

	found := false
	for _, e := range *res.JSON200.Items {
		isFingerprintNull := data.Fingerprint.IsNull()

		if *e.Name == data.Name.ValueString() {
			if isFingerprintNull || *e.Fingerprint == data.Fingerprint.ValueString() {
				found = true

				nData := KeyPairResourceModel{
					Id:          types.StringValue(*e.Name),
					Name:        types.StringValue(*e.Name),
					PrivateKey:  data.PrivateKey,
					PublicKey:   data.PublicKey,
					Fingerprint: types.StringValue(*e.Fingerprint),
				}
				response.Diagnostics.Append(response.State.Set(ctx, &nData)...)
			}
		}
	}

	if !found {
		response.State.RemoveResource(ctx)
	}
}

func (k *KeyPairResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data KeyPairResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	res, err := k.client.DeleteKeyPairWithResponse(ctx, data.Name.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Deleting Key Pair", err.Error())
		return
	}

	numspotError := conns.HandleError(http.StatusNoContent, res.HTTPResponse.StatusCode, res.Body)
	if numspotError != nil {
		response.Diagnostics.AddError(numspotError.Title, numspotError.Detail)
		return
	}
}
