package GoLib

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EnvironmentConfig struct {
	Name             string    `json:"name"`
	ActiveServices   *[]string `json:"services"`
	ActiveSince      *time.Time
	ConfigUpdated    *time.Time
	RoutingKeyOutput string
	RoutingKeyInput  string
	InputQueueName   string
	ServiceName      string
}

type CreateServicePayload struct {
	ImageName string            `json:"image" yaml:"image"`
	Tag       string            `json:"tag,omitempty" yaml:"tag,omitempty"`
	EnvVars   map[string]string `json:"env_vars" yaml:"environment"`
	Networks  []string          `json:"networks" yaml:"networks"`
	Secrets   []string          `json:"secrets" yaml:"secrets"`
	Volumes   map[string]string `json:"volumes" yaml:"-"`
	Ports     map[string]string `json:"ports,omitempty" yaml:"-"`
	Deploy    Deploy            `json:"deploy,omitempty" yaml:"deploy"`
}

type ServiceYAML struct {
	Services map[string]ServiceYAMLDetails `yaml:"services"`
}

type ServiceYAMLDetails struct {
	Image    string            `yaml:"image"`
	Hostname string            `yaml:"hostname"`
	Networks []string          `yaml:"networks"`
	Ports    []string          `yaml:"ports"`
	EnvVars  map[string]string `yaml:"environment"`
	Secrets  []string          `yaml:"secrets"`
	Volumes  []string          `yaml:"volumes"`
	Deploy   Deploy            `yaml:"deploy,omitempty"`
}

type Deploy struct {
	Replicas  int       `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	Placement Placement `json:"placement,omitempty" yaml:"placement,omitempty"`
	Resources Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}

type Placement struct {
	Constraints []string `json:"constraints,omitempty" yaml:"constraints,omitempty"`
}

type Resources struct {
	Reservations Resource `json:"reservations,omitempty" yaml:"reservations,omitempty"`
	Limits       Resource `json:"limits,omitempty" yaml:"limits,omitempty"`
}

type Resource struct {
	Memory string `json:"memory,omitempty" yaml:"memory,omitempty"`
}

type Service struct {
	Services map[string]CreateServicePayload `yaml:"services"`
}

type DetachAttachServicePayload struct {
	ServiceName string `json:"service_name"`
	QueueName   string `json:"queue_name"`
}

type KillServicePayload struct {
	ServiceName string `json:"service_name"`
}

type EtcdClient interface {
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
}

type EtcdClientWrapper struct {
	*clientv3.Client
}

func (w *EtcdClientWrapper) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return w.Client.Put(ctx, key, val, opts...)
}
