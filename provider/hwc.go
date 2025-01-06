package provider

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gookit/goutil"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	bssintl "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/model"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/region"
)

func HWC() {
	// Fetch AK/SK from environment variables for security
	ak := "*"
	sk := "*"

	if ak == "" || sk == "" {
		fmt.Println("Error: CLOUD_SDK_AK or CLOUD_SDK_SK environment variables not set")
		return
	}

	auth := global.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()

	client := bssintl.NewBssintlClient(
		bssintl.BssintlClientBuilder().
			WithRegion(region.ValueOf("ap-southeast-1")).
			WithCredential(auth).
			Build())

	now := time.Now()
	currentMonth := now.Format("2006-01")
	lastMonth := now.AddDate(0, -1, 0).Format("2006-01")

	currentMonthCost, err := fetchMonthlyCost(client, currentMonth)
	if err != nil {
		fmt.Println("Error fetching current month expenditures:", err)
		return
	}

	lastMonthCost, err := fetchMonthlyCost(client, lastMonth)
	if err != nil {
		fmt.Println("Error fetching last month expenditures:", err)
		return
	}

	costData := CostData{
		Granularity: "monthly",
		Currency:    goutil.String(currentMonthCost.Currency),
		MonthToDate: Usage{
			UsageStart: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Unix(),
			UsageEnd:   now.Unix(),
			UsageCost:  currentMonthCost.TotalAmount.InexactFloat64(),
		},
		LastMonth: Usage{
			UsageStart: time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC).Unix(),
			UsageEnd:   time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Unix(),
			UsageCost:  lastMonthCost.TotalAmount.InexactFloat64(),
		},
	}

	// Marshal the response data to JSON
	responseJSON, err := json.MarshalIndent(map[string]CostData{"data": costData}, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling data to JSON:", err)
		return
	}

	fmt.Println(string(responseJSON))
}

func fetchMonthlyCost(client *bssintl.BssintlClient, cycle string) (*model.ListMonthlyExpendituresResponse, error) {
	request := &model.ListMonthlyExpendituresRequest{
		Cycle: cycle,
	}

	response, err := client.ListMonthlyExpenditures(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}
