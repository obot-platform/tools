package commands

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func ListSubnets(ctx context.Context, client *ec2.Client, vpcId string) error {
	var subnets []types.Subnet
	subnetPaginator := ec2.NewDescribeSubnetsPaginator(client, &ec2.DescribeSubnetsInput{})
	for subnetPaginator.HasMorePages() {
		output, err := subnetPaginator.NextPage(ctx)
		if err != nil {
			return err
		} else {
			subnets = append(subnets, output.Subnets...)
		}
	}

	fmt.Printf("Subnets in %s:\n", vpcId)
	for _, s := range subnets {
		fmt.Printf("Subnet Id: %s\n", *s.SubnetId)
		subnetName := "N/A"
		for _, tag := range s.Tags {
			if *tag.Key == "Name" {
				subnetName = *tag.Value
			}
		}
		fmt.Printf("  Name: %s\n", subnetName)
		fmt.Printf("  CIDR: %s\n", *s.CidrBlock)
		fmt.Printf("  State: %s\n", string(s.State))
		fmt.Printf("  Availability Zone: %s\n", *s.AvailabilityZone)
		fmt.Printf("  Available IPs: %d\n", *s.AvailableIpAddressCount)
	}
	return nil
}
