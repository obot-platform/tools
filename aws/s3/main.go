package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/obot-platform/tools/aws/s3/pkg/commands"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: s3 <command>")
	}

	command := os.Args[1]
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	cfg.Region = region

	client := s3.NewFromConfig(cfg)
	bucket := os.Getenv("BUCKET_NAME")

	switch command {
	case "listBuckets":
		if err := commands.ListBuckets(ctx, client); err != nil {
			log.Fatal(err)
		}
	case "listObjects":
		if bucket == "" {
			log.Fatal("bucket_name parameter is required")
		}
		params := commands.ListObjectsParams{
			Bucket: bucket,
		}

		limitString := os.Getenv("LIMIT")
		if limitString == "" {
			limitString = "100"
		}
		limit, err := strconv.Atoi(limitString)
		if err != nil {
			log.Fatal(err)
		}
		params.Limit = limit

		client, err := commands.NewRegionMatchedS3Client(ctx, client, bucket)
		if err != nil {
			log.Fatal(err)
		}
		if err := commands.ListObjects(ctx, client, params); err != nil {
			log.Fatal(err)
		}
	case "getBucketInfo":
		if bucket == "" {
			log.Fatal("bucket_name parameter is required")
		}
		client, err := commands.NewRegionMatchedS3Client(ctx, client, bucket)
		if err != nil {
			log.Fatal(err)
		}
		if err := commands.GetBucketInfo(ctx, client, bucket); err != nil {
			log.Fatal(err)
		}
	case "getObject":
		if bucket == "" {
			log.Fatal("bucket_name parameter is required")
		}
		key := os.Getenv("OBJECT_KEY")
		if key == "" {
			log.Fatal("object_key parameter is required")
		}
		key = strings.TrimPrefix(key, "s3://")
		params := commands.GetObjectParams{
			Bucket: bucket,
			Key:    key,
		}
		client, err := commands.NewRegionMatchedS3Client(ctx, client, bucket)
		if err != nil {
			log.Fatal(err)
		}
		if err := commands.GetObject(ctx, client, params); err != nil {
			log.Fatal(err)
		}
	case "putObject":
		if bucket == "" {
			log.Fatal("bucket_name parameter is required")
		}
		key := os.Getenv("OBJECT_KEY")
		if key == "" {
			log.Fatal("object_key parameter is required")
		}
		localPath := os.Getenv("LOCAL_PATH")
		key = strings.TrimPrefix(key, "s3://")
		params := commands.PutObjectParams{
			Bucket: bucket,
			Key:    key,
		}
		client, err := commands.NewRegionMatchedS3Client(ctx, client, bucket)
		if err != nil {
			log.Fatal(err)
		}

		if err := commands.PutObject(ctx, client, localPath, params); err != nil {
			log.Fatal(err)
		}
	case "deleteObject":
		if bucket == "" {
			log.Fatal("bucket_name parameter is required")
		}
		key := os.Getenv("OBJECT_KEY")
		if key == "" {
			log.Fatal("object_key parameter is required")
		}
		key = strings.TrimPrefix(key, "s3://")
		params := commands.DeleteObjectParams{
			Bucket: bucket,
			Key:    key,
		}
		client, err := commands.NewRegionMatchedS3Client(ctx, client, bucket)
		if err != nil {
			log.Fatal(err)
		}
		if err := commands.DeleteObject(ctx, client, params); err != nil {
			log.Fatal(err)
		}
	}
}
