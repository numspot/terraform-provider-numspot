package core

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/utils"
)

type ListBucketsOutput struct {
	AllBuckets *Buckets `json:"buckets,omitempty" xml:"Buckets"`
}

type Buckets struct {
	Buckets []Bucket `json:"buckets,omitempty" xml:"Bucket"`
}

type Bucket struct {
	CreationDate string `json:"creationDate,omitempty" xml:"CreationDate"`
	Name         string `json:"name,omitempty" xml:"Name"`
}

func CreateBucket(ctx context.Context, provider *client.NumSpotSDK, bucketName string) error {
	_, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	signFn := utils.SignRequest(provider.S3Creds.Service, provider.S3Creds.Region, provider.S3Creds.Ak, provider.S3Creds.Sk)

	res, err := provider.OsClient.CreateBucket(ctx, provider.SpaceID, bucketName, signFn)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("Failed to create Bucket %d \n", res.StatusCode)
	}

	return nil
}

func DeleteBucket(ctx context.Context, provider *client.NumSpotSDK, bucketName string) error {
	_, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	signFn := utils.SignRequest(provider.S3Creds.Service, provider.S3Creds.Region, provider.S3Creds.Ak, provider.S3Creds.Sk)

	res, err := provider.OsClient.DeleteBucket(ctx, provider.SpaceID, bucketName, signFn)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Failed to create Bucket %d\n", res.StatusCode)
	}

	return nil
}

func ReadBucket(ctx context.Context, provider *client.NumSpotSDK, bucketName string) (*Bucket, error) {
	_, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := ReadBuckets(ctx, provider)
	if err != nil {
		return nil, err
	}

	var ret Bucket
	ll := len(res.AllBuckets.Buckets)
	brk := false
	for i := 0; ll > i && !brk; i++ {
		if bucketName == (res.AllBuckets.Buckets)[i].Name {
			brk = true
			ret = (res.AllBuckets.Buckets)[i]
		}
	}

	return &ret, nil
}

func ReadBuckets(ctx context.Context, provider *client.NumSpotSDK) (*ListBucketsOutput, error) {
	_, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	signFn := utils.SignRequest(provider.S3Creds.Service, provider.S3Creds.Region, provider.S3Creds.Ak, provider.S3Creds.Sk)

	res, err := provider.OsClient.ListBuckets(ctx, provider.SpaceID, signFn)
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(res.Body)
	listBucketResponseSchema := ListBucketsOutput{}
	err = decoder.Decode(&listBucketResponseSchema)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &listBucketResponseSchema, nil
}
