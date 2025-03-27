package commands

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"strings"
)

type LaunchInstanceParams struct {
	InstanceType     string
	ImageId          string
	SubnetId         string
	KeyName          string
	SecurityGroupIds []string
	Tags             []string
}

func LaunchInstance(ctx context.Context, client *ec2.Client, params LaunchInstanceParams) error {
	var tags []types.Tag
	for _, tag := range params.Tags {
		kvPair := strings.Split(tag, "=")
		if len(kvPair) != 2 {
			fmt.Printf("-- Skipping Invalid tag format '%s'", tag)
			continue
		}
		tags = append(tags, types.Tag{
			Key:   aws.String(kvPair[0]),
			Value: aws.String(kvPair[1]),
		})
	}
	input := &ec2.RunInstancesInput{
		MaxCount:         aws.Int32(1),
		MinCount:         aws.Int32(1),
		InstanceType:     types.InstanceType(params.InstanceType),
		ImageId:          aws.String(params.ImageId),
		SubnetId:         aws.String(params.SubnetId),
		KeyName:          aws.String(params.KeyName),
		SecurityGroupIds: params.SecurityGroupIds,
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags:         tags,
			},
			{
				ResourceType: types.ResourceTypeVolume,
				Tags:         tags,
			},
			{
				ResourceType: types.ResourceTypeNetworkInterface,
				Tags:         tags,
			},
		},
	}
	instance, err := client.RunInstances(ctx, input)
	if err != nil {
		return err
	}
	fmt.Printf("Launching instance %s\n", *instance.Instances[0].InstanceId)

	return nil
}
