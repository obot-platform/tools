package commands

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awshttp "github.com/aws/smithy-go/transport/http"
)

func GetBucketInfo(ctx context.Context, client *s3.Client, bucket string) error {
	locationInput := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	}
	locationOutput, err := client.GetBucketLocation(ctx, locationInput)
	if err != nil {
		return fmt.Errorf("failed to get bucket location: %v", err)
	}
	region := string(locationOutput.LocationConstraint)
	if region == "" {
		region = "us-east-1"
	}

	var replicationRole string
	replicationInput := &s3.GetBucketReplicationInput{
		Bucket: aws.String(bucket),
	}
	replicationOutput, err := client.GetBucketReplication(ctx, replicationInput)
	if err != nil {
		var respError *awshttp.ResponseError
		if errors.As(err, &respError) && respError.HTTPStatusCode() == http.StatusNotFound {
			replicationRole = "Not Configured"
		} else {
			return fmt.Errorf("failed to get bucket replication: %v", err)
		}
	} else {
		replicationRole = *replicationOutput.ReplicationConfiguration.Role
	}

	var createdDate string
	listInput := &s3.ListBucketsInput{
		Prefix: aws.String(bucket),
	}
	listOutput, err := client.ListBuckets(ctx, listInput)
	if err != nil {
		return err
	}
	for _, bucketInfo := range listOutput.Buckets {
		if *bucketInfo.Name == bucket {
			createdDate = bucketInfo.CreationDate.String()
		}
	}

	var webConfig string
	websiteInput := &s3.GetBucketWebsiteInput{
		Bucket: aws.String(bucket),
	}
	websiteOutput, err := client.GetBucketWebsite(ctx, websiteInput)
	if err != nil {
		var respError *awshttp.ResponseError
		if errors.As(err, &respError) && respError.HTTPStatusCode() == http.StatusNotFound {
			webConfig = "Not Configured"
		} else {
			return fmt.Errorf("failed to get bucket website: %v", err)
		}
	} else {
		webConfig = "Configured"
	}

	fmt.Printf("Bucket: %s\n", bucket)
	fmt.Printf("  Created Date: %s\n", createdDate)
	fmt.Printf("  Region: %s\n", region)
	fmt.Printf("  Replication Role: %s\n", replicationRole)

	fmt.Printf("  Website Configuration: %s\n", webConfig)
	if webConfig != "Not Configured" {
		fmt.Printf("    Index Document: %s\n", *websiteOutput.IndexDocument.Suffix)
		fmt.Printf("    Error Document: %s\n", *websiteOutput.ErrorDocument.Key)
		if len(websiteOutput.RoutingRules) > 0 {
			fmt.Println("    Routing Rules:")
			for _, rule := range websiteOutput.RoutingRules {
				fmt.Printf("    - Condition: %+v\n", rule.Condition)
				fmt.Printf("      Redirect: %+v\n", rule.Redirect)
			}
		}
	}
	return nil
}
