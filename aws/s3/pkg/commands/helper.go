package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

func NewRegionMatchedS3Client(ctx context.Context, client *s3.Client, bucket string) (*s3.Client, error) {
	var err error
	var region string
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err == nil {
		return client, nil
	}

	var respErr *smithyhttp.ResponseError
	if errors.As(err, &respErr) {
		hdr := respErr.Response.Header.Get("x-amz-bucket-region")
		if hdr == "" {
			return nil, fmt.Errorf("failed to find s3 region: %w", err)
		}
		region = hdr
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	cfg.Region = region
	regionalClient := s3.NewFromConfig(cfg)

	return regionalClient, nil
}
