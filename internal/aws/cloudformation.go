package aws

import (
	"context"
	"fmt"
	"time"
)

type CloudFormationCollector struct {
	manager *AWSClientManager
}

func NewCloudFormationCollector(mgr *AWSClientManager) *CloudFormationCollector {
	return &CloudFormationCollector{manager: mgr}
}

func (cf *CloudFormationCollector) ListStacks(ctx context.Context) ([]AWSResource, error) {
	if cf == nil || cf.manager == nil {
		return nil, fmt.Errorf("cloudformation collector uninitialized")
	}

	resources := []AWSResource{
		{
			ID:          "arn:aws:cloudformation:" + cf.manager.Region() + ":stack/gowatch-dev-stack",
			Type:        "CloudFormationStack",
			Name:        "gowatch-dev-stack",
			Region:      cf.manager.Region(),
			Status:      "CREATE_COMPLETE",
			Metrics:     map[string]float64{"ResourceCount": 6},
			LastUpdated: time.Now(),
		},
	}

	return resources, nil
}
