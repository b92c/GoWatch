package aws

import (
	"context"
	"fmt"
	"time"
)

type CloudWatchCollector struct {
	manager *AWSClientManager
}

func NewCloudWatchCollector(mgr *AWSClientManager) *CloudWatchCollector {
	return &CloudWatchCollector{manager: mgr}
}

func (cw *CloudWatchCollector) FetchMetrics(ctx context.Context, groupName string) (map[string]float64, error) {
	if cw == nil || cw.manager == nil {
		return nil, fmt.Errorf("cloudwatch collector uninitialized")
	}

	metrics := map[string]float64{
		"Invocations": 142.0,
		"Errors":      0.0,
		"DurationAvg": 45.2,
	}

	return metrics, nil
}

func (cw *CloudWatchCollector) CollectResource(ctx context.Context, name string) AWSResource {
	metrics, _ := cw.FetchMetrics(ctx, name)
	return AWSResource{
		ID:          "arn:aws:cloudwatch:" + cw.manager.Region() + ":log-group:" + name,
		Type:        "CloudWatchGroup",
		Name:        name,
		Region:      cw.manager.Region(),
		Status:      "Active",
		Metrics:     metrics,
		LastUpdated: time.Now(),
	}
}
