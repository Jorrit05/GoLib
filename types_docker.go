package GoLib

type Service struct {
	Services map[string]CreateServicePayload `yaml:"services"`
}

type MicroServiceData struct {
	Services map[string]MicroServiceDetails `yaml:"services"`
}

type MicroServiceDetails struct {
	Tag     string
	Image   string            `yaml:"image"`
	Ports   map[string]string `yaml:"ports"`
	EnvVars map[string]string `yaml:"environment"`
	Secrets []string          `yaml:"secrets"`
	Volumes map[string]string `yaml:"volumes"`
	Deploy  Deploy            `yaml:"deploy,omitempty"`
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

type ExternalDockerConfig struct {
	Networks []string `yaml:"networks"`
	Volumes  []string `yaml:"volumes"`
	Secrets  []string `yaml:"secrets"`
}
