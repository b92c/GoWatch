package aws

import (
	"context"
	"fmt"
	"time"
)

type XRayCollector struct {
	manager *AWSClientManager
}

func NewXRayCollector(mgr *AWSClientManager) *XRayCollector {
	return &XRayCollector{manager: mgr}
}

func (x *XRayCollector) FetchServiceGraph(ctx context.Context) ([]AWSResource, error) {
	if x == nil || x.manager == nil {
		return nil, fmt.Errorf("xray collector uninitialized")
	}

	resources := []AWSResource{
		{
			ID:          "arn:aws:xray:" + x.manager.Region() + ":service:api-gateway",
			Type:        "XRayService",
			Name:        "api-gateway",
			Region:      x.manager.Region(),
			Status:      "OK",
			Metrics:     map[string]float64{"LatencyMs": 18.4, "OkCount": 540, "ErrorCount": 0},
			LastUpdated: time.Now(),
		},
	}

	return resources, nil
}
