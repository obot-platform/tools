package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"log"
)

type DeleteObjectParams struct {
	Bucket string
	Key    string
}

func DeleteObject(ctx context.Context, client *s3.Client, params DeleteObjectParams) error {
	if params.Bucket == "" {
		return fmt.Errorf("bucket name is required")
	}
	if params.Key == "" {
		return fmt.Errorf("key is required")
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(params.Bucket),
		Key:    aws.String(params.Key),
	}

	_, err := client.DeleteObject(ctx, input)
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't delete object %s from bucket %s. No such key exists.\n", params.Key, params.Bucket)
			err = noKey
		} else {
			log.Printf("Couldn't delete object %v:%v. Here's why: %v\n", params.Bucket, params.Key, err)
		}
		return err
	}
	
	fmt.Printf("Deleted object %s from bucket %s.\n", params.Key, params.Bucket)
	return nil
}
