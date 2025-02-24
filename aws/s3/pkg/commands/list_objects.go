package commands

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type ListObjectsParams struct {
	Bucket string
	Limit  int // 0 means no limit
}

func ListObjects(ctx context.Context, client *s3.Client, params ListObjectsParams) error {
	if params.Bucket == "" {
		return fmt.Errorf("bucket name is required")
	}

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(params.Bucket),
		MaxKeys: aws.Int32(int32(params.Limit)),
	}

	paginator := s3.NewListObjectsV2Paginator(client, input)
	var objects []types.Object
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		if len(objects) >= params.Limit {
			break
		}
		objects = append(objects, output.Contents...)
	}
	fmt.Println("Key -- Last Modified Date")
	for _, obj := range objects {
		fmt.Printf("%s -- %s\n", *obj.Key, *obj.LastModified)
	}

	return nil
}
