package GoLib

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v2"
)

func GetEtcdClient() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"etcd1:2379", "etcd2:2379", "etcd3:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	return cli
}

type leaseOptions struct {
	leaseTime int64
}

type Option func(*leaseOptions)

func LeaseTime(leaseTime int64) Option {
	return func(options *leaseOptions) {
		options.leaseTime = leaseTime
	}
}

// Create object in Etcd with a default 5 second lease
func CreateEtcdLeaseObject(etcdClient *clientv3.Client, key string, value string, opts ...Option) {
	// Default options
	options := &leaseOptions{
		leaseTime: 5,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(options)
	}

	// Create a lease with a 5-second TTL
	lease, err := etcdClient.Grant(context.Background(), options.leaseTime)
	if err != nil {
		log.Fatal(err)
	}

	// Write agent information to etcd with the lease attached
	_, err = etcdClient.Put(context.Background(), key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		log.Fatalf("Failed creating a item with lease in etcd: %s", err)
	}

	// Keep the lease alive by refreshing it periodically
	leaseKeepAlive, err := etcdClient.KeepAlive(context.Background(), lease.ID)
	if err != nil {
		log.Fatalf("Failed starting the keepalive for etcd: %s", err)
	}

	// Periodically refresh the lease
	for range leaseKeepAlive {
		log.Debugf("Lease refreshed on key: %s", key)
	}
}

func UnmarshalStackFile(fileLocation string) MicroServiceData {

	yamlFile, err := os.ReadFile(fileLocation)
	if err != nil {
		log.Errorf("Failed to read the YAML file: %v", err)
	}

	service := MicroServiceData{}
	err = yaml.Unmarshal(yamlFile, &service)
	if err != nil {
		log.Errorf("Failed to unmarshal the YAML file: %v", err)
	}
	return service
}

// Take a given docker stack yaml file, and save all pertinent info (struct MicroServiceData), like the
// required env variable and volumes etc. Into etcd.
func SetMicroservicesEtcd(etcdClient EtcdClient, fileLocation string, etcdPath string) (map[string]MicroServiceDetails, error) {
	if etcdPath == "" {
		etcdPath = "/microservices"
	}

	var service MicroServiceData = UnmarshalStackFile(fileLocation)

	processedServices := make(map[string]MicroServiceDetails)

	for serviceName, payload := range service.Services {

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("Failed to marshal the payload to JSON: %v", err)
			return nil, err
		}

		_, err = etcdClient.Put(context.Background(), fmt.Sprintf("%s/%s", etcdPath, serviceName), string(jsonPayload))
		if err != nil {
			log.Errorf("Failed creating service config in etcd: %s", err)
			return nil, err
		}
		processedServices[serviceName] = payload

	}
	return processedServices, nil
}

func GetValueFromEtcd(etcdClient *clientv3.Client, key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := etcdClient.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get key %s from etcd: %v", key, err)
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key %s not found in etcd", key)
	}

	value := string(resp.Kvs[0].Value)
	return value, nil
}

func GetKeyValueMap(etcdClient *clientv3.Client, pathName string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := etcdClient.Get(ctx, pathName, clientv3.WithPrefix())
	if err != nil {
		log.Errorf("failed to get keys with prefix %s from etcd: %v", pathName, err)
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		log.Errorf("no keys with prefix %s found in etcd", pathName)
		return nil, err
	}

	values := make(map[string]string)
	for _, kv := range resp.Kvs {
		values[string(kv.Key)] = string(kv.Value)
	}
	return values, nil
}

// RegisterJSONArray takes a JSON array, unmarshals it into the target Iterable,
// and stores each element in the etcd key-value store.
//   - T is the underlying struct type of the target.
//   - jsonContent is the byte array containing the JSON content.
//   - target should be an instance of a struct that implements the Iterable and NameGetter interfaces.
//   - etcdClient is an instance of the etcd client.
//   - key is the etcd key prefix where the elements will be stored.
func RegisterJSONArray[T any](jsonContent []byte, target Iterable, etcdClient *clientv3.Client, key string) error {
	err := json.Unmarshal(jsonContent, &target)
	if err != nil {
		log.Errorf("failed to unmarshal JSON content: %v", err)
		return err
	}

	for i := 0; i < target.Len(); i++ {
		element := target.Get(i).(NameGetter) // Assert that element implements NameGetter

		jsonRep, err := json.Marshal(element)
		if err != nil {
			log.Errorf("Failed to Marshal config: %v", err)
			return err
		}

		_, err = etcdClient.Put(context.Background(), fmt.Sprintf("%s/%s", key, string(element.GetName())), string(jsonRep))
		if err != nil {
			log.Errorf("Failed creating archetypesJSON in etcd: %s", err)
			return err
		}
	}

	return nil
}

// GetAndUnmarshalJSON retrieves a JSON value from etcd and unmarshals it into the target struct.
// - T should be a pointer to a struct type.
// - etcdClient is an instance of the etcd client.
// - key is the etcd key where the JSON value is stored.
// - target should be a pointer to an instance of the target struct.
func GetAndUnmarshalJSON[T any](etcdClient *clientv3.Client, key string, target T) error {
	// Get the value from etcd.
	resp, err := etcdClient.Get(context.Background(), key)
	if err != nil {
		log.Errorf("failed to get value from etcd: %v", err)
		return err
	}

	if len(resp.Kvs) == 0 {
		log.Errorf("no value found for key: %s", key)
		return err
	}

	// Unmarshal the JSON value into the target struct.
	err = json.Unmarshal(resp.Kvs[0].Value, target)
	if err != nil {
		log.Errorf("failed to unmarshal JSON: %v", err)
		return err
	}

	return nil
}

// T should be a struct type.
// Pass a full path (like /microservices/) and get a Map back of all entries in that folder.
//
// See etcd_test.go for examples
func GetAndUnmarshalJSONMap[T any](etcdClient *clientv3.Client, prefix string) (map[string]T, error) {
	// Get all key-value pairs under the specified prefix.
	resp, err := etcdClient.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get values from etcd: %v", err)
	}

	// Initialize an empty map to store the unmarshaled structs.
	result := make(map[string]T)

	// Iterate through the key-value pairs and unmarshal the values into structs.
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		// Extract the map key from the etcd key.
		mapKey := strings.TrimPrefix(key, prefix)
		if mapKey == "" {
			continue
		}

		// Unmarshal the JSON value into the target struct.
		var target T
		err = json.Unmarshal(kv.Value, &target)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON for key %s: %v", key, err)
		}

		// Add the unmarshaled struct to the result map.
		result[mapKey] = target
	}

	return result, nil
}

// T is the struct type to be saved.
// target is an instance of the struct.
// etcdClient is an instance of the etcd client.
// key is the etcd key where the value will be stored.
func SaveStructToEtcd[T any](etcdClient *clientv3.Client, key string, target T) error {
	// Marshal the target struct into a JSON representation
	jsonRep, err := json.Marshal(target)
	if err != nil {
		log.Errorf("failed to marshal struct: %v", err)
		return err
	}

	// Save the JSON representation to the etcd key-value store
	_, err = etcdClient.Put(context.Background(), key, string(jsonRep))
	if err != nil {
		log.Errorf("failed to save struct to etcd: %v", err)
		return err
	}

	return nil
}
