package commands

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func ListBuckets(ctx context.Context, client *s3.Client) error {
	var buckets []types.Bucket
	bucketPaginator := s3.NewListBucketsPaginator(client, &s3.ListBucketsInput{})
	for bucketPaginator.HasMorePages() {
		output, err := bucketPaginator.NextPage(ctx)
		if err != nil {
			return err
		} else {
			buckets = append(buckets, output.Buckets...)
		}
	}

	fmt.Println("Buckets:")
	for _, b := range buckets {
		fmt.Println(*b.Name)
	}
	return nil
}
