package GoLib

type ArcheTypes struct {
	ArcheTypes []ArcheType `json:"archetypes"`
}

type ArcheType struct {
	ArcheTypeName string   `json:"name"`
	RequestType   string   `json:"request_type"`
	IoConfig      IoConfig `json:"io_config"`
}

type IoConfig struct {
	ServiceIO      map[string]string `json:"service_io"`
	Finish         string            `json:"finish"`
	ThirdPartyName string            `json:"third_party_name"`
	ThirdParty     map[string]string `json:"third_party"`
}
