package GoLib

type Reasoner struct {
	RequestorId      string   `json:"requestor_id"`
	CurrentArchetype string   `json:"current_archetype"`
	AllowedPartners  []string `json:"allowed_partners"`
}

type ReasonerConfig struct {
	ReasonerConfig []Reasoner `json:"reasoner_config"`
}
