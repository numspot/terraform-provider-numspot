package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_space"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_space"
)

const (
	TfRequestRetryTimeout = 5 * time.Minute
	TfRequestRetryDelay   = 2 * time.Second
)

func SpaceFromTfToCreateRequest(tf *resource_space.SpaceModel) numspot.CreateSpaceRequest {
	return numspot.CreateSpaceRequest{
		Description: tf.Description.ValueString(),
		Name:        tf.Name.ValueString(),
	}
}

func SpaceFromHttpToTf(http *numspot.Space) resource_space.SpaceModel {
	return resource_space.SpaceModel{
		Id:             types.StringValue(http.Id.String()),
		Name:           types.StringValue(http.Name),
		Description:    types.StringValue(http.Description),
		OrganisationId: types.StringValue(http.OrganisationId.String()),
		Status:         types.StringValue(string(http.Status)),
		CreatedOn:      types.StringValue(http.CreatedOn.String()),
		UpdatedOn:      types.StringValue(http.UpdatedOn.String()),
	}
}

func RetryReadSpaceUntilReady(ctx context.Context, client *numspot.ClientWithResponses, spaceID numspot.SpaceId) (interface{}, error) {
	pendingStates := []string{"", "QUEUED", "RUNNING"}
	targetStates := []string{"READY"}
	createStateConf := &retry.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			res, err := client.GetSpaceByIdWithResponse(ctx, spaceID)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read space : %v", err.Error())
			}
			return res.JSON200, string(res.JSON200.Status), nil
		},
		Timeout: TfRequestRetryTimeout,
		Delay:   TfRequestRetryDelay,
	}

	return createStateConf.WaitForStateContext(ctx)
}

func SpaceFromHttpToTfDatasource(ctx context.Context, http *numspot.Space) (*datasource_space.SpaceModel, diag.Diagnostics) {
	return &datasource_space.SpaceModel{
		Id:             types.StringValue(http.Id.String()),
		Name:           types.StringValue(http.Name),
		Description:    types.StringValue(http.Description),
		OrganisationId: types.StringValue(http.OrganisationId.String()),
		Status:         types.StringValue(string(http.Status)),
		CreatedOn:      types.StringValue(http.CreatedOn.String()),
		UpdatedOn:      types.StringValue(http.UpdatedOn.String()),
		SpaceId:        types.StringValue(http.Id.String()),
	}, nil
}
