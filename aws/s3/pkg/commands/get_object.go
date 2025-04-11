package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gptscript-ai/go-gptscript"
	"io"
	"log"
	"path/filepath"
)

type GetObjectParams struct {
	Bucket string
	Key    string
}

func GetObject(ctx context.Context, client *s3.Client, params GetObjectParams) error {
	if params.Bucket == "" {
		return fmt.Errorf("bucket name is required")
	}
	if params.Key == "" {
		return fmt.Errorf("key is required")
	}

	fileName := filepath.Base(params.Key)

	input := &s3.GetObjectInput{
		Bucket: aws.String(params.Bucket),
		Key:    aws.String(params.Key),
	}

	options := func(o *s3.Options) {
		o.DisableLogOutputChecksumValidationSkipped = true
	}
	output, err := client.GetObject(ctx, input, options)
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get object %s from bucket %s. No such key exists.\n", params.Key, params.Bucket)
			err = noKey
		} else {
			log.Printf("Couldn't get object %v:%v. Here's why: %v\n", params.Bucket, params.Key, err)
		}
		return err
	}
	defer output.Body.Close()

	gs, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}
	body, err := io.ReadAll(output.Body)
	if err != nil {
		return err
	}
	err = gs.WriteFileInWorkspace(ctx, "files/"+fileName, body)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %s to the workspace.", fileName)
	return nil
}
