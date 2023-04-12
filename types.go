package GoLib

import "time"

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
	ImageName string            `json:"image_name"`
	Tag       string            `json:"tag"`
	EnvVars   map[string]string `json:"env_vars"`
	Networks  []string          `json:"networks"`
	Secrets   []string          `json:"secrets"`
	Volumes   map[string]string `json:"volumes"`
	Ports     map[string]string `json:"ports"`
}

type DetachAttachServicePayload struct {
	ServiceName string `json:"service_name"`
	QueueName   string `json:"queue_name"`
}

type KillServicePayload struct {
	ServiceName string `json:"service_name"`
}
