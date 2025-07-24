package main

import (
	"context"
	"fmt"
	"os"

	"github.com/obot-platform/tools/aws/costmanagement/pkg/commands"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	fmt.Println("AWS Cost Management Tool")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-2"))
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		os.Exit(1)
	}

	commands.GetCostAndUsage(cfg)
}
