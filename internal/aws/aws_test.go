package aws

import (
	"context"
	"testing"
)

func TestAWSClientManager(t *testing.T) {
	mgr := NewAWSClientManager()

	if mgr.Region() == "" {
		t.Errorf("expected default region, got empty")
	}

	state := mgr.GetState()
	if state.Region != mgr.Region() {
		t.Errorf("expected state region %s, got %s", mgr.Region(), state.Region)
	}

	mgr.SetConnected(true, "")
	if !mgr.GetState().Connected {
		t.Errorf("expected state connected = true")
	}

	resources := []AWSResource{
		{ID: "test-id", Name: "test-resource", Type: "Lambda"},
	}
	mgr.SetResources(resources)

	if len(mgr.GetState().Resources) != 1 {
		t.Fatalf("expected 1 resource in state, got %d", len(mgr.GetState().Resources))
	}
}

func TestAWSCollectors(t *testing.T) {
	ctx := context.Background()
	mgr := NewAWSClientManager()

	cw := NewCloudWatchCollector(mgr)
	res := cw.CollectResource(ctx, "gowatch-logs")
	if res.Type != "CloudWatchGroup" || res.Name != "gowatch-logs" {
		t.Errorf("unexpected cloudwatch resource: %+v", res)
	}

	lambda := NewLambdaCollector(mgr)
	funcs, err := lambda.ListFunctions(ctx)
	if err != nil || len(funcs) == 0 {
		t.Errorf("failed to list lambda functions: %v", err)
	}

	xray := NewXRayCollector(mgr)
	services, err := xray.FetchServiceGraph(ctx)
	if err != nil || len(services) == 0 {
		t.Errorf("failed to fetch xray service graph: %v", err)
	}

	cf := NewCloudFormationCollector(mgr)
	stacks, err := cf.ListStacks(ctx)
	if err != nil || len(stacks) == 0 {
		t.Errorf("failed to list cloudformation stacks: %v", err)
	}
}
