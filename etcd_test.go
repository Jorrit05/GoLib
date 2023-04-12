package GoLib

import (
	"context"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type mockEtcdClient struct {
	data map[string]string
}

func (m *mockEtcdClient) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	m.data[key] = val
	return nil, nil
}

// main_test.go
func TestSetMicroservicesEtcd(t *testing.T) {
	// Mock the etcd client
	mockClient := &mockEtcdClient{
		data: make(map[string]string),
	}

	// Call SetMicroservicesEtcd with the mock client
	processedServices, err := SetMicroservicesEtcd(mockClient, "./microservices_test.yml")
	if err != nil {
		t.Fatalf("Error setting microservices in etcd: %v", err)
	}

	orchestratorPayload := processedServices["anonymize_service"]

	// Check the resulting payload structure for the orchestrator service
	if orchestratorPayload.Image != "anonymize_service" || orchestratorPayload.Tag != "latest" || len(orchestratorPayload.Ports) > 0 {
		t.Errorf("Unexpected payload structure for orchestrator service: %+v", orchestratorPayload)
	}
	// Add more checks for other services if necessary
}
