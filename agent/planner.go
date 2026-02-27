package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hairizuan-noorazman/ui-automation/logger"
)

// Planner is the first agent in the pipeline. It creates an exploration strategy.
type Planner struct {
	config Config
	logger logger.Logger
}

// NewPlanner creates a new planner agent.
func NewPlanner(config Config, log logger.Logger) *Planner {
	return &Planner{
		config: config,
		logger: log,
	}
}

const plannerSystemPrompt = `You are a UI exploration planner. Given a target web application URL and available credentials, create a comprehensive exploration strategy.

Your output must be valid JSON with these fields:
- target_url: The URL to explore
- strategy: A brief description of the exploration approach
- page_areas: List of page areas/sections to explore
- actions: Ordered list of actions to take (navigate, click, type, etc.)
- credentials: Credentials to use for login if needed

Focus on:
1. Identifying the main navigation structure
2. Key interactive elements (forms, buttons, links)
3. Different page states (logged in vs logged out)
4. Edge cases and error states`

// Plan creates an exploration plan for the given target.
func (p *Planner) Plan(ctx context.Context, targetURL string, credentials []Credential) (*ExplorationPlan, error) {
	p.logger.Info(ctx, "creating exploration plan", map[string]interface{}{
		"target_url": targetURL,
	})

	// TODO: Call Claude via Bedrock to generate the plan
	// For now, return a default plan
	plan := &ExplorationPlan{
		TargetURL: targetURL,
		Strategy:  "Navigate to the target URL, explore all visible pages and interactions",
		PageAreas: []string{"homepage", "navigation", "main content"},
		Actions: []string{
			fmt.Sprintf("Navigate to %s", targetURL),
			"Take screenshot of landing page",
			"Identify and click main navigation links",
			"Take screenshots of each page",
		},
		Credentials: credentials,
	}

	planJSON, _ := json.Marshal(plan)
	p.logger.Info(ctx, "exploration plan created", map[string]interface{}{
		"plan": string(planJSON),
	})

	return plan, nil
}
