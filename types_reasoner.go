package GoLib

type Reasoner struct {
	Archetype        string `json:"archetype"`
	RequiredServices []RequiredService
}

type RequiredService struct {
	ServiceName string `json:"service_name"`
}
