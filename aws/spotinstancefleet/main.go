package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/obot-platform/tools/aws/spotinstancefleet/pkg/commands"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: spotinstancefleet <command>")
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

	switch command {
	case "listSpotFleetInstances":
		fleetID := os.Getenv("FLEET_ID")
		if fleetID == "" {
			log.Fatal("fleet_id parameter is required")
		}
		commands.ListSpotFleetInstances(fleetID)
	case "listSpotFleetRequests":
		commands.ListSpotFleetRequests(region) // Pass the region as an argument
	case "setSpotFleetTarget":
		fleetID := os.Getenv("FLEET_ID")
		if fleetID == "" {
			log.Fatal("fleet_id parameter is required")
		}
		targetCapacityStr := os.Getenv("TARGET_CAPACITY")
		if targetCapacityStr == "" {
			log.Fatal("target_capacity parameter is required")
		}
		targetCapacity, err := strconv.Atoi(targetCapacityStr)
		if err != nil {
			log.Fatalf("Invalid target_capacity: %v", err)
		}
		terminateInstances := os.Getenv("TERMINATE_INSTANCES") == "true"
		commands.SetSpotFleetTarget(fleetID, int32(targetCapacity), terminateInstances, region) // Pass the region as an argument
	}
}
