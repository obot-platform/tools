package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func ListSecurityGroups(ctx context.Context, client *ec2.Client, vpcId string) error {
	var securityGroups []types.SecurityGroup
	securityGroupPaginator := ec2.NewDescribeSecurityGroupsPaginator(client, &ec2.DescribeSecurityGroupsInput{})
	for securityGroupPaginator.HasMorePages() {
		output, err := securityGroupPaginator.NextPage(ctx)
		if err != nil {
			return err
		} else {
			securityGroups = append(securityGroups, output.SecurityGroups...)
		}
	}

	fmt.Printf("Security Groups in %s:\n", vpcId)
	for _, s := range securityGroups {
		fmt.Printf("Security Group Id: %s\n", *s.GroupId)
		fmt.Printf("  Name: %s\n", *s.GroupName)
		fmt.Printf("  Description: %s\n", *s.Description)

		b, err := json.MarshalIndent(s.IpPermissions, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("  Ingress rules:\n")
		fmt.Println("```")
		fmt.Println(string(b))
		fmt.Println("```")
		b, err = json.MarshalIndent(s.IpPermissionsEgress, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("  Egress rules:\n")
		fmt.Println("```")
		fmt.Println(string(b))
		fmt.Println("```")
		fmt.Println()
	}
	return nil
}
