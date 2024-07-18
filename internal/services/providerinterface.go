package services

import (
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"
)

type IProvider interface {
	GetSpaceID() numspot.SpaceId
	GetNumspotClient() *numspot.ClientWithResponses
}
