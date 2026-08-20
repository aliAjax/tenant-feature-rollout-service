package domain

type Environment struct {
	ID        string            `json:"id"`
	ProjectID string            `json:"project_id"`
	Name      string            `json:"name"`
	Policies  map[string]string `json:"policies"`
}

func (e Environment) Valid() bool {
	return e.ID != "" && e.ProjectID != "" && e.Name != ""
}

func NewPolicySet(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func (e *Environment) SetPolicy(key, value string) {
	if e.Policies == nil {
		e.Policies = make(map[string]string)
	}
	e.Policies[key] = value
}

func (e Environment) Clone() Environment {
	e.Policies = NewPolicySet(e.Policies)
	return e
}
