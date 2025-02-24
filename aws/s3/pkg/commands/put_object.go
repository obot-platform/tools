package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/gptscript-ai/go-gptscript"
	"log"
	"path/filepath"
	"time"
)

type PutObjectParams struct {
	Bucket string
	Key    string
}

func PutObject(ctx context.Context, client *s3.Client, filePath string, params PutObjectParams) error {
	if params.Bucket == "" {
		return fmt.Errorf("bucket name is required")
	}
	if params.Key == "" {
		return fmt.Errorf("key is required")
	}
	if filePath == "" {
		return fmt.Errorf("file path is required")
	}

	gs, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}
	body, err := gs.ReadFileInWorkspace(ctx, filepath.Join("files", filePath))
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(body)

	input := &s3.PutObjectInput{
		Bucket: aws.String(params.Bucket),
		Key:    aws.String(params.Key),
		Body:   bodyReader,
	}

	options := func(o *s3.Options) {
		o.DisableLogOutputChecksumValidationSkipped = true
	}
	_, err = client.PutObject(ctx, input, options)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			log.Printf("Cannot upload an object larger than 5GB.", params.Bucket)
		} else {
			log.Printf("Couldn't upload file to %s:%s. Here's why: %s\n", params.Bucket, params.Key, err)
		}
	} else {
		err = s3.NewObjectExistsWaiter(client).Wait(
			ctx, &s3.HeadObjectInput{Bucket: aws.String(params.Bucket), Key: aws.String(params.Key)}, time.Minute)
		if err != nil {
			log.Printf("Failed attempt to wait for object %s to exist.\n", params.Key)
		}
	}
	fmt.Printf("Successfully uploaded file to s3://%s/%s.\n", params.Bucket, params.Key)
	return nil
}
