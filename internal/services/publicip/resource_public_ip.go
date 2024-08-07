package publicip

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &PublicIpResource{}
	_ resource.ResourceWithConfigure   = &PublicIpResource{}
	_ resource.ResourceWithImportState = &PublicIpResource{}
)

type PublicIpResource struct {
	provider services.IProvider
}

func NewPublicIpResource() resource.Resource {
	return &PublicIpResource{}
}

func (r *PublicIpResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(services.IProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *PublicIpResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *PublicIpResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_public_ip"
}

func (r *PublicIpResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = PublicIpResourceSchema(ctx)
}

func (r *PublicIpResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan PublicIpModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	createRes, err := utils2.RetryCreateUntilResourceAvailable(
		ctx,
		r.provider.GetSpaceID(),
		r.provider.GetNumspotClient().CreatePublicIpWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Public IP", err.Error())
		return
	}

	publicIp, diags := PublicIpFromHttpToTf(ctx, createRes.JSON201)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	createdId := *createRes.JSON201.Id
	if len(plan.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, plan.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Attach the public IP to a VM or NIC if their IDs are provided:
	// Note: According to the resource schema, vmId and nicId cannot be set simultaneously.
	// This constraint is enforced by the stringvalidator.ConflictsWith function.
	if (!plan.VmId.IsNull() && !plan.VmId.IsUnknown()) || (!plan.NicId.IsNull() && !plan.NicId.IsUnknown()) {
		publicIp.VmId = plan.VmId
		publicIp.NicId = plan.NicId

		// Call Link publicIP
		linkPublicIP, err := invokeLinkPublicIP(ctx, r.provider, publicIp)
		if err != nil {
			response.Diagnostics.AddError("Failed to link public IP", err.Error())
		}
		publicIp.LinkPublicIP = types.StringPointerValue(linkPublicIP)
	}

	// Refresh state
	data, diags := refreshState(ctx, r.provider, publicIp.Id.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, *data)...)
}

func (r *PublicIpResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data PublicIpModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	readRes := utils2.ExecuteRequest(func() (*numspot.ReadPublicIpsByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadPublicIpsByIdWithResponse(ctx, r.provider.GetSpaceID(), data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if readRes == nil {
		return
	}

	tf, diags := PublicIpFromHttpToTf(ctx, readRes.JSON200)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *PublicIpResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		plan, state  PublicIpModel
		linkPublicIP *string
		err          error
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			state.Id.ValueString(),
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	data, diags := refreshState(ctx, r.provider, state.Id.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	chgSet := ComputePublicIPChangeSet(&plan, data)

	if chgSet.Err != nil {
		response.Diagnostics.AddError("Failed to update public IP", chgSet.Err.Error())
		return
	}

	if chgSet.Unlink {
		_ = invokeUnlinkPublicIP(ctx, r.provider, &state) // We still want to try delete resource even if the unlink didn't work (ressource has been unlinked before for example)
		state.LinkPublicIP = types.StringNull()
		data, diags := refreshState(ctx, r.provider, state.Id.ValueString())
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		response.Diagnostics.Append(response.State.Set(ctx, *data)...)
	}
	if chgSet.Link {
		plan.Id = state.Id
		linkPublicIP, err = invokeLinkPublicIP(ctx, r.provider, &plan)
		if err != nil {
			response.Diagnostics.AddError("Failed to link public IP", err.Error())
			return
		}
		state.LinkPublicIP = types.StringPointerValue(linkPublicIP)
	}
	data, diags = refreshState(ctx, r.provider, state.Id.ValueString())
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, *data)...)
}

func (r *PublicIpResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state PublicIpModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if !state.LinkPublicIP.IsNull() {
		_ = invokeUnlinkPublicIP(ctx, r.provider, &state) // We still want to try delete resource even if the unlink didn't work (ressource has been unlinked before for example)
	}

	err := utils2.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), state.Id.ValueString(), r.provider.GetNumspotClient().DeletePublicIpWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Public IP", err.Error())
		return
	}
}
