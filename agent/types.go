package agent

// ExplorationPlan is the output of the planner agent.
type ExplorationPlan struct {
	TargetURL   string       `json:"target_url"`
	Strategy    string       `json:"strategy"`
	PageAreas   []string     `json:"page_areas"`
	Actions     []string     `json:"actions"`
	Credentials []Credential `json:"credentials"`
}

type Credential struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Interaction represents a single browser interaction during exploration.
type Interaction struct {
	StepNumber     int    `json:"step_number"`
	ActionType     string `json:"action_type"`
	Description    string `json:"description"`
	ScreenshotPath string `json:"screenshot_path,omitempty"`
}

// ExplorationResult is the output of the explorer agent.
type ExplorationResult struct {
	Interactions []Interaction `json:"interactions"`
	Summary      string        `json:"summary"`
}
