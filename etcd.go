package GoLib

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// Take a given docker stack yaml file, and save all pertinent info (struct MicroServiceData), like the
// required env variable and volumes etc. Into etcd.
func SetMicroservicesEtcd(etcdClient EtcdClient, fileLocation string) (map[string]MicroServiceDetails, error) {
	yamlFile, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		log.Errorf("Failed to read the YAML file: %v", err)
	}

	service := MicroServiceData{}
	err = yaml.Unmarshal(yamlFile, &service)
	if err != nil {
		log.Errorf("Failed to unmarshal the YAML file: %v", err)
	}

	processedServices := make(map[string]MicroServiceDetails)

	for serviceName, payload := range service.Services {

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("Failed to marshal the payload to JSON: %v", err)
			return nil, err
		}

		_, err = etcdClient.Put(context.Background(), fmt.Sprintf("/microservices/%s", serviceName), string(jsonPayload))
		if err != nil {
			log.Errorf("Failed creating service config in etcd: %s", err)
			return nil, err
		}
		processedServices[serviceName] = payload

	}
	return processedServices, nil
}
