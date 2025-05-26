package commands

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// ListSpotFleetRequests lists all Spot Fleet requests
// If region is empty, it will use the default region from AWS config
func ListSpotFleetRequests(region string) {
	fmt.Println("Loading AWS configuration...")
	opts := []func(*config.LoadOptions) error{
		config.WithSharedConfigProfile("mihai_skylitetek_root"),
	}

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		fmt.Println("Error loading configuration", err)
		return
	}

	// Afișăm regiunea folosită pentru debugging
	fmt.Printf("Using AWS region: %s\n", cfg.Region)

	fmt.Println("Creating EC2 service client...")
	svc := ec2.NewFromConfig(cfg)

	input := &ec2.DescribeSpotFleetRequestsInput{}

	fmt.Println("Describing Spot Fleet requests...")
	result, err := svc.DescribeSpotFleetRequests(context.TODO(), input)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Printf("Found %d Spot Fleet requests.\n", len(result.SpotFleetRequestConfigs))
	for _, request := range result.SpotFleetRequestConfigs {
		config := request.SpotFleetRequestConfig
		fmt.Printf("- Spot Fleet Request ID: %s\n", *request.SpotFleetRequestId)
		fmt.Printf("- State: %s\n", request.SpotFleetRequestState)
		fmt.Printf("- Activity Status: %s\n", request.ActivityStatus)
		fmt.Printf("- Create Time: %s\n", request.CreateTime)
		fmt.Printf("- Allocation Strategy: %s\n", config.AllocationStrategy)
		if config.OnDemandAllocationStrategy != "" {
			fmt.Printf("- On-Demand Allocation Strategy: %s\n", config.OnDemandAllocationStrategy)
		} else {
			fmt.Println("- On-Demand Allocation Strategy: N/A")
		}
		fmt.Printf("- IAM Fleet Role: %s\n", *config.IamFleetRole)
		fmt.Printf("- Fulfilled Capacity: %f\n", *config.FulfilledCapacity)
		fmt.Printf("- On-Demand Fulfilled Capacity: %f\n", *config.OnDemandFulfilledCapacity)
		fmt.Printf("- Target Capacity: %d\n", *config.TargetCapacity)
		fmt.Printf("- On-Demand Target Capacity: %d\n", *config.OnDemandTargetCapacity)
		fmt.Printf("- Terminate Instances With Expiration: %t\n", *config.TerminateInstancesWithExpiration)
		fmt.Printf("- Type: %s\n", config.Type)
		fmt.Printf("- Valid From: %s\n", config.ValidFrom)
		fmt.Printf("- Valid Until: %s\n", config.ValidUntil)
		fmt.Printf("- Replace Unhealthy Instances: %t\n", *config.ReplaceUnhealthyInstances)
		fmt.Printf("- Instance Interruption Behavior: %s\n", config.InstanceInterruptionBehavior)

		fmt.Println("\n### Launch Specifications:")
		for _, spec := range config.LaunchSpecifications {
			fmt.Printf("- Image ID: %s\n", *spec.ImageId)
			fmt.Printf("- Instance Type: %s\n", spec.InstanceType)
			fmt.Printf("- Key Name: %s\n", *spec.KeyName)
			fmt.Printf("- Monitoring: %t\n", *spec.Monitoring.Enabled)
			for _, ni := range spec.NetworkInterfaces {
				fmt.Printf("- Network Interface: Associate Public IP Address: %t, Delete On Termination: %t\n", *ni.AssociatePublicIpAddress, *ni.DeleteOnTermination)
				fmt.Printf("- Subnet ID: %s\n", *ni.SubnetId)
			}
			for _, tagSpec := range spec.TagSpecifications {
				for _, tag := range tagSpec.Tags {
					fmt.Printf("- Tag: %s - %s\n", *tag.Key, *tag.Value)
				}
			}
		}
		fmt.Println("\n")
	}
}
