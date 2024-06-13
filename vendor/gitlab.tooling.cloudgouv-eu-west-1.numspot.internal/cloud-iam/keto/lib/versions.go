package lib

import (
	"context"
	"fmt"

	relations "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
)

// GetVersion returns the server's running version of Keto.
func (client *readClient) GetVersion(ctx context.Context) (string, error) {
	req := relations.GetVersionRequest{}
	res, err := client.versionClient.GetVersion(ctx, &req)
	if err != nil {
		return "", fmt.Errorf("grpc call: %w", err)
	}

	return res.GetVersion(), nil
}
