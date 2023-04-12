package GoLib

type ExternalDockerConfig struct {
	Networks []string `yaml:"networks"`
	Volumes  []string `yaml:"volumes"`
	Secrets  []string `yaml:"secrets"`
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
