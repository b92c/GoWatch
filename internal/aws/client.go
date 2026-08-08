package aws

import (
	"net/http"
	"os"
	"sync"
	"time"
)

type AWSResource struct {
	Metrics     map[string]float64 `json:"metrics,omitempty"`
	LastUpdated time.Time          `json:"last_updated"`
	ID          string             `json:"id"`
	Type        string             `json:"type"`
	Name        string             `json:"name"`
	Region      string             `json:"region"`
	Status      string             `json:"status"`
}

type AWSState struct {
	Resources   []AWSResource `json:"resources"`
	Region      string        `json:"region"`
	EndpointURL string        `json:"endpoint_url,omitempty"`
	LastError   string        `json:"last_error,omitempty"`
	Enabled     bool          `json:"enabled"`
	Connected   bool          `json:"connected"`
}

type AWSClientManager struct {
	httpClient  *http.Client
	region      string
	endpointURL string
	state       AWSState
	mu          sync.RWMutex
}

func NewAWSClientManager() *AWSClientManager {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = "us-east-1"
	}

	endpointURL := os.Getenv("AWS_ENDPOINT_URL")

	enabled := os.Getenv("AWS_ACCESS_KEY_ID") != "" || os.Getenv("AWS_PROFILE") != "" || endpointURL != "" || os.Getenv("AWS_REGION") != ""

	return &AWSClientManager{
		region:      region,
		endpointURL: endpointURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		state: AWSState{
			Enabled:     enabled,
			Region:      region,
			EndpointURL: endpointURL,
			Resources:   make([]AWSResource, 0),
		},
	}
}

func (m *AWSClientManager) GetState() AWSState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resCopy := make([]AWSResource, len(m.state.Resources))
	copy(resCopy, m.state.Resources)
	st := m.state
	st.Resources = resCopy
	return st
}

func (m *AWSClientManager) SetConnected(connected bool, errStr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Connected = connected
	m.state.LastError = errStr
}

func (m *AWSClientManager) SetResources(resources []AWSResource) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Resources = resources
	m.state.Connected = true
	m.state.LastError = ""
}

func (m *AWSClientManager) Region() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.region
}

func (m *AWSClientManager) EndpointURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.endpointURL
}
