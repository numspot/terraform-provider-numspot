package core

import (
	"context"
	"fmt"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateVM(ctx context.Context, provider *client.NumSpotSDK, numSpotVolumeCreate numspot.CreateVolumeJSONRequestBody, tags []numspot.ResourceTag, vmID, deviceName string) (numSpotVM *numspot.Vm, err error) {

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		VmFromTfToCreateRequest(ctx, &data, &response.Diagnostics),
		numspotClient.CreateVmsWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", err.Error())
	}
	if response.Diagnostics.HasError() {
		return
	}

	vm := *res.JSON201
	createdId := *vm.Id

	// Create tags
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, numspotClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	read, err := utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"running", "stopped"}, // In some cases, when there is insufficient capacity the VM is created with state = stopped
		numspotClient.ReadVmsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", createdId, err))
		return
	}

	vmSchema, ok := read.(*numspot.Vm)
	if !ok {
		response.Diagnostics.AddError("Failed to create VM", "object conversion error")
		return
	}

	// In some cases, when there is insufficient capacity the VM is created with state = stopped
	if utils.GetPtrValue(vmSchema.State) == "stopped" {
		response.Diagnostics.AddError("Issue while creating VM", fmt.Sprintf("VM was created in 'stopped' state. Reason : %s", utils.GetPtrValue(vmSchema.StateReason)))
		return
	}

}
