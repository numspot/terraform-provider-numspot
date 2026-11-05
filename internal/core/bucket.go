package core

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	// errFailedToCreateBucket is returned when the object-storage create-bucket
	// call returns an unexpected status code.
	errFailedToCreateBucket = errors.New("failed to create Bucket")
	// errFailedToDeleteBucket is returned when the object-storage delete-bucket
	// call returns an unexpected status code.
	errFailedToDeleteBucket = errors.New("failed to delete Bucket")
	// errFailedToListBuckets is returned when the object-storage list-buckets
	// call returns an unexpected status code.
	errFailedToListBuckets = errors.New("failed to list buckets")
)

// ListBucketsOutput handles list buckets output.
type ListBucketsOutput struct {
	AllBuckets *Buckets `json:"buckets,omitempty" xml:"Buckets"`
}

// Buckets handles buckets.
type Buckets struct {
	Buckets []Bucket `json:"buckets,omitempty" xml:"Bucket"`
}

// Bucket handles bucket.
type Bucket struct {
	CreationDate string `json:"creationDate,omitempty" xml:"CreationDate"`
	Name         string `json:"name,omitempty" xml:"Name"`
}

// CreateBucket creates bucket.
func CreateBucket(ctx context.Context, provider *client.NumspotSDK, bucketName string) error {
	_, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	res, err := provider.OsClient.CreateBucket(ctx, provider.SpaceID, bucketName, provider.SignFunc)
	if err != nil {
		return utils.WrapError(err, "failed to create bucket")
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("%w %d", errFailedToCreateBucket, res.StatusCode)
	}

	return nil
}

// DeleteBucket deletes bucket.
func DeleteBucket(ctx context.Context, provider *client.NumspotSDK, bucketName string) error {
	_, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	res, err := provider.OsClient.DeleteBucket(ctx, provider.SpaceID, bucketName, provider.SignFunc)
	if err != nil {
		return utils.WrapError(err, "failed to delete bucket")
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%w %d", errFailedToDeleteBucket, res.StatusCode)
	}

	return nil
}

// ReadBucket reads bucket.
func ReadBucket(ctx context.Context, provider *client.NumspotSDK, bucketName string) (*Bucket, error) {
	_, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := ReadBuckets(ctx, provider)
	if err != nil {
		return nil, err
	}

	var ret Bucket
	ll := len(res.AllBuckets.Buckets)
	brk := false
	for i := 0; ll > i && !brk; i++ {
		if bucketName == res.AllBuckets.Buckets[i].Name {
			brk = true
			ret = res.AllBuckets.Buckets[i]
		}
	}

	return &ret, nil
}

// ReadBuckets reads buckets.
func ReadBuckets(ctx context.Context, provider *client.NumspotSDK) (*ListBucketsOutput, error) {
	_, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	res, err := provider.OsClient.ListBuckets(ctx, provider.SpaceID, provider.SignFunc)
	if err != nil {
		return nil, utils.WrapError(err, "failed to list buckets")
	}
	defer func() { _ = res.Body.Close() }()

	if http.StatusOK != res.StatusCode {
		return nil, fmt.Errorf("%w: %d", errFailedToListBuckets, res.StatusCode)
	}

	decoder := xml.NewDecoder(res.Body)
	listBucketResponseSchema := ListBucketsOutput{}
	err = decoder.Decode(&listBucketResponseSchema)
	if err != nil && err != io.EOF {
		return nil, utils.WrapError(err, "failed to decode bucket list XML")
	}

	return &listBucketResponseSchema, nil
}
