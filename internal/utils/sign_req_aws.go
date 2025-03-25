package utils

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	awsv4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"terraform-provider-numspot/internal/sdk/objectstorage"
)

func SignRequest(service, region, accessKey, secretKey string) objectstorage.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		staticCreds := credentials.NewStaticCredentials(accessKey, secretKey, "")

		signer := awsv4.NewSigner(staticCreds)

		_, err := signer.Sign(req, nil, service, region, time.Now())
		if err != nil {
			return err
		}

		return nil
	}
}
