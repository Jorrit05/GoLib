package GoLib

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
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

// Create object in Etcd with a default 5 second lease
func CreateEtcdLeaseObject(etcdClient *clientv3.Client, key string, value string, leaseTime int64) {
	var actualLeaseTime int64 = leaseTime
	if leaseTime == 0 {
		actualLeaseTime = 5
	}
	// Create a lease with a 5-second TTL
	lease, err := etcdClient.Grant(context.Background(), actualLeaseTime)
	if err != nil {
		log.Fatal(err)
	}

	// Write agent information to etcd with the lease attached
	_, err = etcdClient.Put(context.Background(), key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		log.Fatal(err)
	}

	// Keep the lease alive by refreshing it periodically
	leaseKeepAlive, err := etcdClient.KeepAlive(context.Background(), lease.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Periodically refresh the lease
	for range leaseKeepAlive {
		log.Debugf("Lease refreshed on key: %s", key)
	}
}
