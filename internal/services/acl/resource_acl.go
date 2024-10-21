package acl

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/serviceaccount"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource              = &AclsResource{}
	_ resource.ResourceWithConfigure = &AclsResource{}
)

type AclsResource struct {
	provider *client.NumSpotSDK
}

func NewAclsResource() resource.Resource {
	return &AclsResource{}
}

func (r *AclsResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	r.provider = provider
}

func (r *AclsResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_acls"
}

func (r *AclsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = ACLsResourceSchema(ctx)
}

func (r *AclsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan ACLsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Attach ACLs
	if len(plan.ACLs.Elements()) > 0 {
		acls := CreateAclListFromTf(ctx, plan, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		r.updateAcls(ctx, serviceaccount.AddAction, plan.SpaceId.ValueString(), plan.ServiceAccountId.ValueString(), acls, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	tf := r.readAcls(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *AclsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ACLsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	tf := r.readAcls(ctx, state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *AclsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("nothing to do")
}

func (r *AclsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state ACLsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	stateAcls := CreateAclListFromTf(ctx, state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	if len(stateAcls) > 0 {
		r.updateAcls(ctx, serviceaccount.DeleteAction, state.SpaceId.ValueString(), state.ServiceAccountId.ValueString(), stateAcls, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (r *AclsResource) updateAcls(
	ctx context.Context,
	action serviceaccount.Action,
	spaceId string,
	serviceAccountID string,
	acls []numspot.ACL,
	diags *diag.Diagnostics,
) {
	spaceUUID := utils.ParseUUID(spaceId, diags)
	if diags.HasError() {
		return
	}

	// Parse Service Account ID
	serviceAccountUUID := utils.ParseUUID(serviceAccountID, diags)
	if diags.HasError() {
		return
	}

	body := numspot.ACLList{
		Items: acls,
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	// Execute
	if action == serviceaccount.AddAction {
		utils.ExecuteRequest(func() (*numspot.CreateACLServiceAccountSpaceBulkResponse, error) {
			return numspotClient.CreateACLServiceAccountSpaceBulkWithResponse(
				ctx,
				spaceUUID,
				serviceAccountUUID,
				body,
			)
		}, http.StatusCreated, diags)
	} else if action == serviceaccount.DeleteAction {
		utils.ExecuteRequest(func() (*numspot.DeleteACLServiceAccountSpaceBulkResponse, error) {
			return numspotClient.DeleteACLServiceAccountSpaceBulkWithResponse(
				ctx,
				spaceUUID,
				serviceAccountUUID,
				body,
			)
		}, http.StatusNoContent, diags)
	}
}

func (r *AclsResource) readAcls(
	ctx context.Context,
	tf ACLsModel,
	diags *diag.Diagnostics,
) *ACLsModel {
	readResult := ACLsModel{
		SpaceId:          types.StringValue(r.provider.SpaceID.String()),
		ServiceAccountId: tf.ServiceAccountId,
		Service:          tf.Service,
		Resource:         tf.Resource,
		Subresource:      tf.Subresource,
	}

	serviceAccountUUID := utils.ParseUUID(tf.ServiceAccountId.ValueString(), diags)
	if diags.HasError() {
		return nil
	}

	body := numspot.GetACLServiceAccountSpaceParams{
		Service:     tf.Service.ValueString(),
		Resource:    tf.Resource.ValueString(),
		Subresource: tf.Subresource.ValueStringPointer(),
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}

	res := utils.ExecuteRequest(func() (*numspot.GetACLServiceAccountSpaceResponse, error) {
		return numspotClient.GetACLServiceAccountSpaceWithResponse(
			ctx, r.provider.SpaceID, serviceAccountUUID, &body)
	}, http.StatusOK, diags)
	if res == nil {
		return nil
	}

	if res.JSON200 == nil {
		diags.AddError("Failed to get IAM ACLs space response", res.Status())
		return nil
	}

	acls := res.JSON200.Items

	if diags.HasError() {
		return nil
	}

	aclsTf := utils.GenericSetToTfSetValue(ctx, CreateTfAclFromHttp, acls, diags)
	if diags.HasError() {
		return nil
	}
	readResult.ACLs = aclsTf

	return &readResult
}
