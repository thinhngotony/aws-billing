package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

type AwsSecret struct {
	SecretAccessKey string `json:"secret_access_key"`
	AccessKeyID     string `json:"access_key_id"`
}

type CostData struct {
	Granularity string `json:"granularity"`
	Currency    string `json:"currency"`
	MonthToDate Usage  `json:"month_to_date"`
	Forecast    Usage  `json:"forecast"`
	LastMonth   Usage  `json:"last_month"`
}

type Usage struct {
	UsageStart int64   `json:"usage_start"`
	UsageEnd   int64   `json:"usage_end"`
	UsageCost  float64 `json:"usage_cost"`
}

func AWS() {
	// Load AWS configuration from secret store
	awsSecret := AwsSecret{
		AccessKeyID:     "*",
		SecretAccessKey: "*",
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsSecret.AccessKeyID, awsSecret.SecretAccessKey, "")),
	)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		return
	}

	client := costexplorer.NewFromConfig(cfg)

	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)
	lastMonthComparisonStart := lastMonthStart
	lastMonthComparisonEnd := time.Date(lastMonthStart.Year(), lastMonthStart.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.UTC)

	// Get current month-to-date cost
	currentCost, err := getCost(client, currentMonthStart.Format("2006-01-02"), now.Format("2006-01-02"))
	if err != nil {
		fmt.Printf("Error getting current month cost: %v\n", err)
		return
	}

	// // Get last month's cost for the same period
	// lastMonthCost, err := getCost(client,
	// 	lastMonthComparisonStart.Format("2006-01-02"),
	// 	lastMonthComparisonEnd.Format("2006-01-02"))
	// if err != nil {
	// 	fmt.Printf("Error getting last month cost: %v\n", err)
	// 	return
	// }

	// Get last month's total cost
	lastMonthTotalCost, err := getCost(client, lastMonthStart.Format("2006-01-02"), currentMonthStart.Format("2006-01-02"))
	if err != nil {
		fmt.Printf("Error getting last month's total cost: %v\n", err)
		return
	}

	// Get forecasted cost for current month
	forecastedCost, err := getForecast(client, now.Format("2006-01-02"), currentMonthStart.AddDate(0, 1, 0).Format("2006-01-02"))
	if err != nil {
		fmt.Printf("Error getting forecasted cost: %v\n", err)
		return
	}

	// Prepare the response data
	costData := CostData{
		Granularity: "monthly",
		Currency:    "usd",
		MonthToDate: Usage{
			UsageStart: currentMonthStart.Unix(),
			UsageEnd:   now.Unix(),
			UsageCost:  parseFloat(currentCost),
		},
		Forecast: Usage{
			UsageStart: now.Unix(),
			UsageEnd:   currentMonthStart.AddDate(0, 1, 0).Unix(),
			UsageCost:  parseFloat(forecastedCost),
		},
		LastMonth: Usage{
			UsageStart: lastMonthComparisonStart.Unix(),
			UsageEnd:   lastMonthComparisonEnd.Unix(),
			UsageCost:  parseFloat(lastMonthTotalCost),
		},
	}

	// Marshal the response data to JSON
	response, err := json.MarshalIndent(map[string]CostData{"data": costData}, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}

	// Print the JSON response
	fmt.Println(string(response))
}

func getCost(client *costexplorer.Client, startDate, endDate string) (string, error) {
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
	}

	result, err := client.GetCostAndUsage(context.TODO(), input)
	if err != nil {
		return "0", err
	}

	var totalCost float64
	for _, period := range result.ResultsByTime {
		if amount := period.Total["UnblendedCost"].Amount; amount != nil {
			totalCost += parseFloat(*amount)
		}
	}

	return fmt.Sprintf("%.2f", totalCost), nil
}

func getForecast(client *costexplorer.Client, startDate, endDate string) (string, error) {
	input := &costexplorer.GetCostForecastInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Granularity: types.GranularityMonthly,
		Metric:      types.MetricUnblendedCost,
	}

	result, err := client.GetCostForecast(context.TODO(), input)
	if err != nil {
		return "0", err
	}

	if result.Total != nil && result.Total.Amount != nil {
		return *result.Total.Amount, nil
	}

	return "0", nil
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
