package utils

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"terraform-provider-numspot/internal/sdk/objectstorage"
)

// emptyStringSHA256 is the SHA-256 hex digest of an empty request body, used
// by the AWS SigV4 signer for object-storage requests that have no payload.
const emptyStringSHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// SignRequest returns a RequestEditorFn that signs outgoing object-storage
// requests with AWS Signature Version 4 using the supplied static credentials.
func SignRequest(service, region, accessKey, secretKey string) objectstorage.RequestEditorFn {
	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""))
	signer := v4.NewSigner()

	return func(ctx context.Context, req *http.Request) error {
		credsValue, err := creds.Retrieve(ctx)
		if err != nil {
			return WrapError(err, "failed to retrieve AWS credentials")
		}

		if err = signer.SignHTTP(ctx, credsValue, req, emptyStringSHA256, service, region, time.Now()); err != nil {
			return WrapError(err, "failed to sign AWS request")
		}

		return nil
	}
}
