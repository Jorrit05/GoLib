package GoLib

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// Take a given docker stack yaml file, and save all pertinent info, like the
// required env variable and volumes etc. Into etcd.
func SetMicroservicesEtcd(etcdClient EtcdClient, fileLocation string) (map[string]CreateServicePayload, error) {
	yamlFile, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		log.Errorf("Failed to read the YAML file: %v", err)
	}

	service := ServiceYAML{}
	err = yaml.Unmarshal(yamlFile, &service)
	if err != nil {
		log.Errorf("Failed to unmarshal the YAML file: %v", err)
	}

	processedServices := make(map[string]CreateServicePayload)

	for serviceName, serviceDetails := range service.Services {
		imageName, tag := SplitImageAndTag(serviceDetails.Image)

		// Volumes and ports are lists in YAML (and json), but need to be maps
		// for working with the dockerspec
		volumes := make(map[string]string)
		for _, volume := range serviceDetails.Volumes {
			parts := strings.Split(volume, ":")
			if len(parts) == 2 {
				volumes[parts[0]] = parts[1]
			}
		}

		ports := make(map[string]string)
		for _, port := range serviceDetails.Ports {
			parts := strings.Split(port, ":")
			if len(parts) == 2 {
				ports[parts[0]] = parts[1]
			}
		}

		payload := CreateServicePayload{
			ImageName: imageName,
			Tag:       tag,
			EnvVars:   serviceDetails.EnvVars,
			Networks:  serviceDetails.Networks,
			Secrets:   serviceDetails.Secrets,
			Volumes:   volumes,
			Ports:     ports,
			Deploy:    serviceDetails.Deploy,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("Failed to marshal the payload to JSON: %v", err)
		}

		_, err = etcdClient.Put(context.Background(), fmt.Sprintf("/microservices/%s", serviceName), string(jsonPayload))
		if err != nil {
			log.Errorf("Failed creating service config in etcd: %s", err)
		}
		processedServices[serviceName] = payload

	}
	return processedServices, nil
}
