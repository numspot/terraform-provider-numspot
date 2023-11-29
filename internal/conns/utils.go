package conns

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/common/slice"
	conns "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api_client"
	"net/http"
)

func HandleError(expectedStatusCode int, response *http.Response) *conns.Error {
	if expectedStatusCode != response.StatusCode {
		var numspotError conns.Error
		defer response.Body.Close()
		err := json.NewDecoder(response.Body).Decode(&numspotError)
		fmt.Println(err)

		return &numspotError
	}

	return nil
}

func GetClient(request resource.ConfigureRequest, response *resource.ConfigureResponse) *conns.ClientWithResponses {
	if request.ProviderData == nil {
		return nil
	}

	client, ok := request.ProviderData.(*conns.ClientWithResponses)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return nil
	}

	return client
}

// TODO: to refactor.
func HandleErrorBis(expectedStatusCode int, responseStatusCode int, responseBody []byte) *conns.Error {
	if expectedStatusCode != responseStatusCode {
		var numSpotError conns.Error
		err := json.Unmarshal(responseBody, &numSpotError)
		fmt.Println(err)

		return &numSpotError
	}

	return nil
}

func MapHttpListToModelList[A, B any](ctx context.Context, httpItems []A, mapperFunc func(A) B, modelObjectType types.ObjectType) (basetypes.ListValue, diag.Diagnostics) {
	modelItems := slice.Map(httpItems, mapperFunc)
	return types.ListValueFrom(ctx, modelObjectType, modelItems)
}
