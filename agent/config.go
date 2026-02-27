package agent

import (
	"time"
)

// Config holds the agent pipeline configuration.
type Config struct {
	MaxIterations    int
	TimeLimit        time.Duration
	BedrockRegion    string
	BedrockModel     string
	PlaywrightMCPURL string
}
