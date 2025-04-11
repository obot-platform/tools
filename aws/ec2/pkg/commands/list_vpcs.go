package commands

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func ListVPCs(ctx context.Context, client *ec2.Client) error {
	var vpcs []types.Vpc
	vpcPaginator := ec2.NewDescribeVpcsPaginator(client, &ec2.DescribeVpcsInput{})
	for vpcPaginator.HasMorePages() {
		output, err := vpcPaginator.NextPage(ctx)
		if err != nil {
			return err
		} else {
			vpcs = append(vpcs, output.Vpcs...)
		}
	}

	fmt.Println("VPCs:")
	for _, v := range vpcs {
		fmt.Printf("VPC Id: %s\n", *v.VpcId)
		vpcName := "N/A"
		for _, tag := range v.Tags {
			if *tag.Key == "Name" {
				vpcName = *tag.Value
			}
		}
		fmt.Printf("  Name: %s\n", vpcName)
		fmt.Printf("  CIDR: %s\n", *v.CidrBlock)
		fmt.Printf("  Is Default: %t\n", *v.IsDefault)
		fmt.Printf("  State: %s\n", string(v.State))
	}
	return nil
}
