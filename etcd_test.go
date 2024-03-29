package lib

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
	processedServices, err := SetMicroservicesEtcd(mockClient, "./microservices_test.yaml", "")
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

func TestGetAndUnmarshalJSON(t *testing.T) {
	cli := setupEtcdClient(t)
	defer cli.Close()

	// Insert test data into etcd
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := cli.Put(ctx, "/microservices/test-service", `{"tag": "test-tag", "image": "test-image", "ports": {"8080": "80"}, "environment": {"VAR": "value"}, "secrets": ["test-secret"], "volumes": {"data": "/data"}, "deploy": {"replicas": 1}}`)
	require.NoError(t, err)

	// Test GetMicroServiceData
	var msData MicroServiceDetails
	_, err = GetAndUnmarshalJSON(cli, "/microservices/test-service", &msData)
	require.NoError(t, err)
	assert.NotNil(t, msData)
	assert.Equal(t, "test-tag", msData.Tag)
	assert.Equal(t, "test-image", msData.Image)

	// Clean up test data from etcd
	_, err = cli.Delete(ctx, "/microservices/test-service")
	require.NoError(t, err)
}
func TestGetAndUnmarshalJSONMap(t *testing.T) {
	cli := setupEtcdClient(t)
	defer cli.Close()

	// Insert test data into etcd
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := cli.Put(ctx, "/testmicroservices/test-service1", `{"tag": "test-tag1", "image": "test-image1", "ports": {"8081": "80"}, "environment": {"VAR1": "value1"}, "secrets": ["test-secret1"], "volumes": {"data1": "/data1"}, "deploy": {"replicas": 1}}`)
	require.NoError(t, err)
	_, err = cli.Put(ctx, "/testmicroservices/test-service2", `{"tag": "test-tag2", "image": "test-image2", "ports": {"8082": "80"}, "environment": {"VAR2": "value2"}, "secrets": ["test-secret2"], "volumes": {"data2": "/data2"}, "deploy": {"replicas": 1}}`)
	require.NoError(t, err)

	// Test GetAndUnmarshalJSONMap
	msDataMap, err := GetAndUnmarshalJSONMap[MicroServiceDetails](cli, "/testmicroservices/")
	require.NoError(t, err)
	assert.NotNil(t, msDataMap)
	assert.Equal(t, 2, len(msDataMap))

	assert.Equal(t, "test-tag1", msDataMap["test-service1"].Tag)
	assert.Equal(t, "test-image1", msDataMap["test-service1"].Image)

	assert.Equal(t, "test-tag2", msDataMap["test-service2"].Tag)
	assert.Equal(t, "test-image2", msDataMap["test-service2"].Image)

	// Clean up test data from etcd
	_, err = cli.Delete(ctx, "/testmicroservices/test-service1")
	require.NoError(t, err)
	_, err = cli.Delete(ctx, "/testmicroservices/test-service2")
	require.NoError(t, err)
}
