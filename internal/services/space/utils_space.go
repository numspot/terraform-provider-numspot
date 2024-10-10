package space

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func SpaceFromTfToCreateRequest(tf *SpaceModel) numspot.CreateSpaceRequest {
	return numspot.CreateSpaceRequest{
		Description: tf.Description.ValueString(),
		Name:        tf.Name.ValueString(),
	}
}

func SpaceFromHttpToTf(http *numspot.Space) SpaceModel {
	return SpaceModel{
		Id:             types.StringValue(http.Id.String()),
		Name:           types.StringValue(http.Name),
		Description:    types.StringValue(http.Description),
		SpaceId:        types.StringValue(http.Id.String()),
		OrganisationId: types.StringValue(http.OrganisationId.String()),
		Status:         types.StringValue(string(http.Status)),
		CreatedOn:      types.StringValue(http.CreatedOn.String()),
		UpdatedOn:      types.StringValue(http.UpdatedOn.String()),
	}
}

func RetryReadSpaceUntilReady(ctx context.Context, client *numspot.ClientWithResponses, organisationID numspot.OrganisationId, spaceID numspot.SpaceId) (interface{}, error) {
	pendingStates := []string{"", "QUEUED", "RUNNING"}
	targetStates := []string{"READY"}
	createStateConf := &retry.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			res, err := client.GetSpaceByIdWithResponse(ctx, organisationID, spaceID)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read space : %v", err.Error())
			}
			// TODO : check that response is a status 200
			return res.JSON200, string(res.JSON200.Status), nil
		},
		Timeout: utils.TfRequestRetryTimeout,
		Delay:   utils.ParseRetryBackoff(),
	}

	return createStateConf.WaitForStateContext(ctx)
}

func SpaceFromHttpToTfDatasource(ctx context.Context, http *numspot.Space) *SpaceModelDataSource {
	return &SpaceModelDataSource{
		Id:             types.StringValue(http.Id.String()),
		Name:           types.StringValue(http.Name),
		Description:    types.StringValue(http.Description),
		OrganisationId: types.StringValue(http.OrganisationId.String()),
		Status:         types.StringValue(string(http.Status)),
		CreatedOn:      types.StringValue(http.CreatedOn.String()),
		UpdatedOn:      types.StringValue(http.UpdatedOn.String()),
		SpaceId:        types.StringValue(http.Id.String()),
	}
}
