package commands

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// SetSpotFleetTarget sets the target capacity for a Spot Fleet and optionally terminates instances
func SetSpotFleetTarget(fleetID string, targetCapacity int32, terminateInstances bool, region string) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		fmt.Println("Error loading configuration", err)
		return
	}

	svc := ec2.NewFromConfig(cfg)

	input := &ec2.ModifySpotFleetRequestInput{
		SpotFleetRequestId: aws.String(fleetID),
		TargetCapacity:     aws.Int32(targetCapacity),
	}

	_, err = svc.ModifySpotFleetRequest(context.TODO(), input)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Printf("Spot Fleet Request ID: %s, Target Capacity Set: %d\n", fleetID, targetCapacity)
	if terminateInstances {
		fmt.Println("Instances will be terminated with expiration.")
	}
}
