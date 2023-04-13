package GoLib

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func setupEtcdClient(t *testing.T) *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	require.NoError(t, err)
	return cli
}

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
	processedServices, err := SetMicroservicesEtcd(mockClient, "./microservices_test.yml", "")
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

func TestGetMicroServiceData(t *testing.T) {
	cli := setupEtcdClient(t)
	defer cli.Close()

	// Insert test data into etcd
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := cli.Put(ctx, "/microservices/test-service", `{"tag": "test-tag", "image": "test-image", "ports": {"8080": "80"}, "environment": {"VAR": "value"}, "secrets": ["test-secret"], "volumes": {"data": "/data"}, "deploy": {"replicas": 1}}`)
	require.NoError(t, err)

	// Test GetMicroServiceData
	msData, err := GetMicroServiceData(cli)
	require.NoError(t, err)
	assert.NotNil(t, msData.Services["test-service"])
	assert.Equal(t, "test-tag", msData.Services["test-service"].Tag)
	assert.Equal(t, "test-image", msData.Services["test-service"].Image)

	// Clean up test data from etcd
	_, err = cli.Delete(ctx, "/microservices/test-service")
	require.NoError(t, err)
}

func TestGetAvailableAgents(t *testing.T) {
	cli := setupEtcdClient(t)
	defer cli.Close()

	// Insert test data into etcd
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := cli.Put(ctx, "/agents/test-agent", `{"name": "test-agent", "services": ["test-service"], "activeSince": "2023-01-01T12:00:00Z", "routingKeyOutput": "test-routing-key-output", "routingKeyInput": "test-routing-key-input", "inputQueueName": "test-input-queue-name", "serviceName": "test-service-name"}`)
	require.NoError(t, err)

	// Test GetAvailableAgents
	agentData, err := GetAvailableAgents(cli)
	require.NoError(t, err)
	assert.NotNil(t, agentData.Agents["test-agent"])
	assert.Equal(t, "test-agent", agentData.Agents["test-agent"].Name)

	// Clean up test data from etcd
	_, err = cli.Delete(ctx, "/agents/test-agent")

	require.NoError(t, err)
}
