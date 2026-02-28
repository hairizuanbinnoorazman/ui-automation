package agent

// AgentConfig is the JSON config sent to the Python agent script via stdin.
type AgentConfig struct {
	TargetURL       string       `json:"target_url"`
	Credentials     []Credential `json:"credentials,omitempty"`
	ProcedureName   string       `json:"procedure_name"`
	JobID           string       `json:"job_id"`
	OutputDir       string       `json:"output_dir"`
	PlaywrightMCPURL string      `json:"playwright_mcp_url"`
}

// Credential holds a key-value pair for endpoint credentials.
type Credential struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AgentResult is the JSON result produced by the Python agent script.
type AgentResult struct {
	ProcedureName string      `json:"procedure_name"`
	Description   string      `json:"description"`
	Steps         []AgentStep `json:"steps"`
	Summary       string      `json:"summary"`
}

// AgentStep represents a single step in the agent-generated test procedure.
type AgentStep struct {
	Name         string   `json:"name"`
	Instructions string   `json:"instructions"`
	ImagePaths   []string `json:"image_paths"`
}
