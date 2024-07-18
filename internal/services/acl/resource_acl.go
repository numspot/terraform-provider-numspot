package acl

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/serviceaccount"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource              = &AclsResource{}
	_ resource.ResourceWithConfigure = &AclsResource{}
)

type AclsResource struct {
	provider services.IProvider
}

func NewAclsResource() resource.Resource {
	return &AclsResource{}
}

func (r *AclsResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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
		acls, diags := CreateAclListFromTf(ctx, plan)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		diags = r.updateAcls(ctx, serviceaccount.AddAction, plan.SpaceId.ValueString(), plan.ServiceAccountId.ValueString(), acls)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
	}

	tf, diags := r.readAcls(ctx, plan)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *AclsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ACLsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	tf, diags := r.readAcls(ctx, state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
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

	stateAcls, diags := CreateAclListFromTf(ctx, state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	if len(stateAcls) > 0 {
		diags := r.updateAcls(ctx, serviceaccount.DeleteAction, state.SpaceId.ValueString(), state.ServiceAccountId.ValueString(), stateAcls)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
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
) diag.Diagnostics {
	var diags diag.Diagnostics

	spaceUUID, diags := utils2.ParseUUID(spaceId)
	if diags.HasError() {
		return diags
	}

	// Parse Service Account ID
	serviceAccountUUID, diags := utils2.ParseUUID(serviceAccountID)
	if diags.HasError() {
		return diags
	}

	body := numspot.ACLList{
		Items: acls,
	}

	if diags.HasError() {
		return diags
	}

	// Execute
	if action == serviceaccount.AddAction {
		utils2.ExecuteRequest(func() (*numspot.CreateACLServiceAccountSpaceBulkResponse, error) {
			return r.provider.GetNumspotClient().CreateACLServiceAccountSpaceBulkWithResponse(
				ctx,
				spaceUUID,
				serviceAccountUUID,
				body,
			)
		}, http.StatusCreated, &diags)
	} else if action == serviceaccount.DeleteAction {
		utils2.ExecuteRequest(func() (*numspot.DeleteACLServiceAccountSpaceBulkResponse, error) {
			return r.provider.GetNumspotClient().DeleteACLServiceAccountSpaceBulkWithResponse(
				ctx,
				spaceUUID,
				serviceAccountUUID,
				body,
			)
		}, http.StatusNoContent, &diags)
	}

	return diags
}

func (r *AclsResource) readAcls(
	ctx context.Context,
	tf ACLsModel,
) (*ACLsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	readResult := ACLsModel{
		SpaceId:          types.StringValue(r.provider.GetSpaceID().String()),
		ServiceAccountId: tf.ServiceAccountId,
		Service:          tf.Service,
		Resource:         tf.Resource,
		Subresource:      tf.Subresource,
	}

	serviceAccountUUID, diags := utils2.ParseUUID(tf.ServiceAccountId.ValueString())
	if diags.HasError() {
		return nil, diags
	}

	body := numspot.GetACLServiceAccountSpaceParams{
		Service:     tf.Service.ValueString(),
		Resource:    tf.Resource.ValueString(),
		Subresource: tf.Subresource.ValueStringPointer(),
	}

	res := utils2.ExecuteRequest(func() (*numspot.GetACLServiceAccountSpaceResponse, error) {
		return r.provider.GetNumspotClient().GetACLServiceAccountSpaceWithResponse(
			ctx, r.provider.GetSpaceID(), serviceAccountUUID, &body)
	}, http.StatusOK, &diags)
	if res == nil {
		return nil, diags
	}

	if res.JSON200 == nil {
		diags.AddError("Failed to get IAM ACLs space response", res.Status())
		return nil, diags
	}

	acls := res.JSON200.Items

	if diags.HasError() {
		return nil, diags
	}

	aclsTf, diags := utils2.GenericSetToTfSetValue(ctx, ACLValue{}, CreateTfAclFromHttp, acls)
	if diags.HasError() {
		return nil, diags
	}
	readResult.ACLs = aclsTf

	return &readResult, nil
}
