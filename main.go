package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

func main() {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		return
	}

	// Create a Cost Explorer client
	client := costexplorer.NewFromConfig(cfg)

	// Get current time
	now := time.Now()

	// Get the first day of current month
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Get the first day of last month
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)

	// Calculate the same period last month
	daysInCurrentPeriod := now.Sub(currentMonthStart).Hours() / 24
	lastMonthSamePeriodEnd := lastMonthStart.AddDate(0, 0, int(daysInCurrentPeriod))

	// Get current month-to-date cost
	currentCost, err := getCost(client, currentMonthStart.Format("2006-01-02"), now.Format("2006-01-02"))
	if err != nil {
		fmt.Printf("Error getting current month cost: %v\n", err)
		return
	}

	// Get last month's cost for the same period
	lastMonthCost, err := getCost(client, lastMonthStart.Format("2006-01-02"), lastMonthSamePeriodEnd.Format("2006-01-02"))
	if err != nil {
		fmt.Printf("Error getting last month cost: %v\n", err)
		return
	}

	// Get last month's total cost
	lastMonthTotalCost, err := getCost(client, lastMonthStart.Format("2006-01-02"), currentMonthStart.Format("2006-01-02"))
	if err != nil {
		fmt.Printf("Error getting last month's total cost: %v\n", err)
		return
	}

	// Calculate the percentage change
	currentAmount := parseFloat(currentCost)
	lastAmount := parseFloat(lastMonthCost)
	percentageChange := ((currentAmount - lastAmount) / lastAmount) * 100

	fmt.Printf("\nCost Summary\n")
	fmt.Printf("Month-to-date cost: $%.2f\n", currentAmount)
	fmt.Printf("%.0f%% compared to last month for same period ($%.2f)\n", percentageChange, lastAmount)
	fmt.Printf("Last month's total cost: $%.2f\n", parseFloat(lastMonthTotalCost))
}

func getCost(client *costexplorer.Client, startDate, endDate string) (string, error) {
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},
	}

	result, err := client.GetCostAndUsage(context.TODO(), input)
	if err != nil {
		return "0", err
	}

	if len(result.ResultsByTime) > 0 {
		return *result.ResultsByTime[0].Total["UnblendedCost"].Amount, nil
	}

	return "0", nil
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
