package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iam"

	resource_acls "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_acl"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource              = &AclsResource{}
	_ resource.ResourceWithConfigure = &AclsResource{}
)

type AclsResource struct {
	provider Provider
}

func NewAclsResource() resource.Resource {
	return &AclsResource{}
}

func (r *AclsResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(Provider)
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
	response.Schema = resource_acls.ACLsResourceSchema(ctx)
}

func (r *AclsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_acls.ACLsModel
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

		diags = r.updateAcls(ctx, AddAction, plan.SpaceId.ValueString(), plan.ServiceAccountId.ValueString(), acls)
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
	var state resource_acls.ACLsModel
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
	var state resource_acls.ACLsModel
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
		diags := r.updateAcls(ctx, DeleteAction, state.SpaceId.ValueString(), state.ServiceAccountId.ValueString(), stateAcls)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (r *AclsResource) updateAcls(
	ctx context.Context,
	action Action,
	spaceId string,
	serviceAccountID string,
	acls []iam.ACL,
) diag.Diagnostics {
	var diags diag.Diagnostics

	spaceUUID, diags := utils.ParseUUID(spaceId)
	if diags.HasError() {
		return diags
	}

	// Parse Service Account ID
	serviceAccountUUID, diags := utils.ParseUUID(serviceAccountID)
	if diags.HasError() {
		return diags
	}

	body := iam.ACLList{
		Items: acls,
	}

	if diags.HasError() {
		return diags
	}

	// Execute
	if action == AddAction {
		utils.ExecuteRequest(func() (*iam.CreateACLServiceAccountSpaceBulkResponse, error) {
			return r.provider.IAMAccessManagerClient.CreateACLServiceAccountSpaceBulkWithResponse(
				ctx,
				spaceUUID,
				serviceAccountUUID,
				body,
			)
		}, http.StatusCreated, &diags)
	} else if action == DeleteAction {
		utils.ExecuteRequest(func() (*iam.DeleteACLServiceAccountSpaceBulkResponse, error) {
			return r.provider.IAMAccessManagerClient.DeleteACLServiceAccountSpaceBulkWithResponse(
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
	tf resource_acls.ACLsModel,
) (*resource_acls.ACLsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	readResult := resource_acls.ACLsModel{
		SpaceId:          types.StringValue(r.provider.SpaceID.String()),
		ServiceAccountId: tf.ServiceAccountId,
		Service:          tf.Service,
		Resource:         tf.Resource,
		Subresource:      tf.Subresource,
	}

	serviceAccountUUID, diags := utils.ParseUUID(tf.ServiceAccountId.ValueString())
	if diags.HasError() {
		return nil, diags
	}

	body := iam.GetACLServiceAccountSpaceParams{
		Service:     tf.Service.ValueString(),
		Resource:    tf.Resource.ValueString(),
		Subresource: tf.Subresource.ValueStringPointer(),
	}

	res := utils.ExecuteRequest(func() (*iam.GetACLServiceAccountSpaceResponse, error) {
		return r.provider.IAMAccessManagerClient.GetACLServiceAccountSpaceWithResponse(
			ctx, r.provider.SpaceID, serviceAccountUUID, &body)
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

	aclsTf, diags := utils.GenericSetToTfSetValue(ctx, resource_acls.ACLValue{}, CreateTfAclFromHttp, acls)
	if diags.HasError() {
		return nil, diags
	}
	readResult.ACLs = aclsTf

	return &readResult, nil
}
