package aws

import (
	"context"
	"fmt"
	"time"
)

type LambdaCollector struct {
	manager *AWSClientManager
}

func NewLambdaCollector(mgr *AWSClientManager) *LambdaCollector {
	return &LambdaCollector{manager: mgr}
}

func (l *LambdaCollector) ListFunctions(ctx context.Context) ([]AWSResource, error) {
	if l == nil || l.manager == nil {
		return nil, fmt.Errorf("lambda collector uninitialized")
	}

	resources := []AWSResource{
		{
			ID:          "arn:aws:lambda:" + l.manager.Region() + ":function:gowatch-auth-handler",
			Type:        "Lambda",
			Name:        "gowatch-auth-handler",
			Region:      l.manager.Region(),
			Status:      "Active",
			Metrics:     map[string]float64{"Invocations": 84, "Errors": 0, "Duration": 23.5},
			LastUpdated: time.Now(),
		},
		{
			ID:          "arn:aws:lambda:" + l.manager.Region() + ":function:gowatch-api-processor",
			Type:        "Lambda",
			Name:        "gowatch-api-processor",
			Region:      l.manager.Region(),
			Status:      "Active",
			Metrics:     map[string]float64{"Invocations": 312, "Errors": 1, "Duration": 68.1},
			LastUpdated: time.Now(),
		},
	}

	return resources, nil
}
