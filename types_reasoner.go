package GoLib

type Requestor struct {
	RequestorId      string   `json:"requestor_id"`
	CurrentArchetype string   `json:"current_archetype"`
	AllowedPartners  []string `json:"allowed_partners"`
}

type RequestorConfig struct {
	ReasonerConfig []Requestor `json:"requestor_config"`
}
