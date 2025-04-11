package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/obot-platform/tools/aws/ec2/pkg/commands"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: ec2 <command>")
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

	client := ec2.NewFromConfig(cfg)

	instanceId := os.Getenv("INSTANCE_ID")
	vpcId := os.Getenv("VPC_ID")

	switch command {
	case "listInstances":
		if err := commands.ListInstances(ctx, client); err != nil {
			log.Fatal(err)
		}
	case "getInstanceInfo":
		if instanceId == "" {
			log.Fatal("instance_id parameter is required")
		}
		if err := commands.GetInstanceInfo(ctx, client, instanceId); err != nil {
			log.Fatal(err)
		}
	case "startInstance":
		if err := commands.ChangeInstanceState(ctx, client, instanceId, "start"); err != nil {
			log.Fatal(err)
		}
	case "stopInstance":
		if err := commands.ChangeInstanceState(ctx, client, instanceId, "stop"); err != nil {
			log.Fatal(err)
		}
	case "rebootInstance":
		if err := commands.ChangeInstanceState(ctx, client, instanceId, "reboot"); err != nil {
			log.Fatal(err)
		}
	case "terminateInstance":
		if err := commands.ChangeInstanceState(ctx, client, instanceId, "terminate"); err != nil {
			log.Fatal(err)
		}
	case "launchInstance":
		instanceType := os.Getenv("INSTANCE_TYPE")
		if instanceType == "" {
			instanceType = "t2.micro"
		}
		imageId := os.Getenv("IMAGE_ID")
		if imageId == "" {
			log.Fatal("image_id parameter is required")
		}
		subnetId := os.Getenv("SUBNET_ID")
		if subnetId == "" {
			log.Fatal("subnet_id parameter is required")
		}
		keyName := os.Getenv("KEY_NAME")
		if keyName == "" {
			log.Fatal("key_name parameter is required")
		}
		stringSecurityGroupIds := os.Getenv("SECURITY_GROUP_IDS")
		if stringSecurityGroupIds == "" {
			log.Fatal("security_group_ids parameter is required")
		}
		securityGroupIds := strings.Split(stringSecurityGroupIds, ",")
		for index, sg := range securityGroupIds {
			securityGroupIds[index] = strings.TrimSpace(sg)
		}
		stringTags := os.Getenv("TAGS")
		tags := strings.Split(stringTags, ",")
		for index, tag := range tags {
			tags[index] = strings.TrimSpace(tag)
		}

		params := commands.LaunchInstanceParams{
			InstanceType:     instanceType,
			ImageId:          imageId,
			SubnetId:         subnetId,
			KeyName:          keyName,
			SecurityGroupIds: securityGroupIds,
			Tags:             tags,
		}

		if err := commands.LaunchInstance(ctx, client, params); err != nil {
			log.Fatal(err)
		}
	case "listVPCs":
		if err := commands.ListVPCs(ctx, client); err != nil {
			log.Fatal(err)
		}
	case "listSubnets":
		if vpcId == "" {
			log.Fatal("vpc_id parameter is required")
		}
		if err := commands.ListSubnets(ctx, client, vpcId); err != nil {
			log.Fatal(err)
		}
	case "listSecurityGroups":
		if vpcId == "" {
			log.Fatal("vpc_id parameter is required")
		}
		if err := commands.ListSecurityGroups(ctx, client, vpcId); err != nil {
			log.Fatal(err)
		}
	case "searchInstancesByName":
		instanceName := os.Getenv("INSTANCE_NAME")
		if instanceName == "" {
			log.Fatal("instance_name parameter is required")
		}
		if err := commands.SearchInstancesByTag(ctx, client, "Name", instanceName); err != nil {
			log.Fatal(err)
		}
	case "searchInstancesByTag":
		tag_key := os.Getenv("TAG_KEY")
		if tag_key == "" {
			log.Fatal("tag_key parameter is required")
		}
		tag_value := os.Getenv("TAG_VALUE")
		if tag_value == "" {
			log.Fatal("tag_value parameter is required")
		}
		if err := commands.SearchInstancesByTag(ctx, client, tag_key, tag_value); err != nil {
			log.Fatal(err)
		}
	}
}
