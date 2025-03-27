package commands

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func SearchInstancesByTag(ctx context.Context, client *ec2.Client, tag_key, tag_value string) error {
	var instances []types.Reservation
	instancePaginator := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", tag_key)),
				Values: []string{tag_value},
			},
		},
	})
	for instancePaginator.HasMorePages() {
		output, err := instancePaginator.NextPage(ctx)
		if err != nil {
			return err
		} else {
			instances = append(instances, output.Reservations...)
		}
	}

	fmt.Println("Matching Instances:")
	for _, i := range instances {
		instance := i.Instances[0]

		fmt.Printf("Instance Id: %s\n", *instance.InstanceId)
		instanceName := "N/A"
		for _, tag := range instance.Tags {
			if *tag.Key == "Name" {
				instanceName = *tag.Value
			}
		}
		fmt.Printf("  Name: %s\n", instanceName)
		fmt.Printf("  VPC Id: %s\n", *instance.VpcId)
		fmt.Printf("  Instance Type: %s\n", instance.InstanceType)
		fmt.Printf("  ImageId: %s\n", *instance.ImageId)
		fmt.Printf("  State: %s\n", instance.State.Name)
	}
	return nil
}
