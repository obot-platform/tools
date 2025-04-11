package commands

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func GetInstanceInfo(ctx context.Context, client *ec2.Client, instanceId string) error {
	params := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-id"),
				Values: []string{instanceId},
			},
		},
	}
	instance, err := client.DescribeInstances(ctx, params)
	if err != nil {
		return err
	}
	if len(instance.Reservations) == 0 || len(instance.Reservations[0].Instances) == 0 {
		return fmt.Errorf("no instances found for instance ID %s", instanceId)
	}
	instanceDetails := instance.Reservations[0].Instances[0]

	fmt.Printf("Instance Id: %s\n", *instanceDetails.InstanceId)
	instanceName := "N/A"
	var tags []string
	for _, tag := range instanceDetails.Tags {
		if *tag.Key == "Name" {
			instanceName = *tag.Value
		}
		tags = append(tags, fmt.Sprintf("%s=%s", *tag.Key, *tag.Value))
	}
	fmt.Printf("  Name: %s\n", instanceName)
	fmt.Printf("  VPC Id: %s\n", *instanceDetails.VpcId)
	fmt.Printf("  Instance Type: %s\n", instanceDetails.InstanceType)
	fmt.Printf("  ImageId: %s\n", *instanceDetails.ImageId)
	fmt.Printf("  State: %s\n", instanceDetails.State.Name)
	publicIp := "N/A"
	if instanceDetails.PublicIpAddress != nil {
		publicIp = *instanceDetails.PublicIpAddress
	}
	fmt.Printf("  Public Ip: %s\n", publicIp)
	fmt.Printf("  Private Ip: %s\n", *instanceDetails.PrivateIpAddress)
	fmt.Printf("  Subnet Id: %s\n", *instanceDetails.SubnetId)
	keyName := "N/A"
	if instanceDetails.KeyName != nil {
		keyName = *instanceDetails.KeyName
	}
	fmt.Printf("  SSH Key Name: %s\n", keyName)
	fmt.Printf("  Launch Time: %s\n", *instanceDetails.LaunchTime)
	instanceLifecycle := "on-demand"
	if instanceDetails.InstanceLifecycle != "" {
		instanceLifecycle = string(instanceDetails.InstanceLifecycle)
	}
	fmt.Printf("  Lifecycle: %s\n", instanceLifecycle)

	fmt.Println("  Security Groups:")
	for _, sg := range instanceDetails.SecurityGroups {
		fmt.Printf("    - \"%s\" (%s)\n", *sg.GroupName, *sg.GroupId)
	}

	fmt.Println("  Tags:")
	for _, tag := range tags {
		fmt.Printf("    - \"%s\"\n", tag)
	}

	return nil
}
