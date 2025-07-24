package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

func GetCostAndUsage(cfg aws.Config) {
	fmt.Println("Starting GetCostAndUsage operation...")
	fmt.Printf("Using AWS region: %s\n", cfg.Region)

	client := costexplorer.NewFromConfig(cfg)

	// Definim perioada de timp (ultimele 30 de zile)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate.Format("2006-01-02")),
			End:   aws.String(endDate.Format("2006-01-02")),
		},
		Granularity: types.GranularityMonthly,
		Metrics: []string{
			"UnblendedCost",
			"UsageQuantity",
		},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
		},
	}

	result, err := client.GetCostAndUsage(context.TODO(), input)
	if err != nil {
		fmt.Printf("Error fetching cost and usage: %v\n", err)
		return
	}

	// Procesăm rezultatele
	for _, result := range result.ResultsByTime {
		fmt.Printf("\nTime Period: %s to %s\n", *result.TimePeriod.Start, *result.TimePeriod.End)
		for _, group := range result.Groups {
			fmt.Printf("\nService: %s\n", group.Keys[0])
			for metricName, metricValue := range group.Metrics {
				fmt.Printf("  %s: %s %s\n", metricName, *metricValue.Amount, *metricValue.Unit)
			}
		}
	}
}
