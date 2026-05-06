package core

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"agentai-go/internal/config"
	"agentai-go/internal/types"
)

// AICore handles Gemini API calls via HTTP with retry and rate limiting.
type AICore struct {
	config *config.Config
	model  string
	inited bool
	lastAt time.Time
}

// NewAICore creates a new AICore.
func NewAICore(cfg *config.Config) *AICore {
	model := cfg.Model
	if model == "" {
		model = cfg.GeminiModel
	}
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &AICore{config: cfg, model: model}
}

func (a *AICore) sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func (a *AICore) ensureRateLimit() {
	elapsed := time.Since(a.lastAt)
	delay := a.config.RequestDelayMs
	if delay <= 0 {
		delay = 2000
	}
	if elapsed < time.Duration(delay)*time.Millisecond {
		a.sleep(delay - int(elapsed.Milliseconds()))
	}
}

// extractRetryDelay tries to get retry delay in ms from error message.
func extractRetryDelay(errStr string) int {
	re := regexp.MustCompile(`retry(?:\s+in|\s*[Dd]elay["\s:]+)?\s*"?(\d+(?:\.\d+)?)\s*s?`)
	if m := re.FindStringSubmatch(errStr); len(m) > 1 {
		sec := 5.0
		fmt.Sscanf(m[1], "%f", &sec)
		return int(sec * 1000)
	}
	return 5000
}

// generateContent calls the configured provider and returns the response text.
func (a *AICore) generateContent(ctx context.Context, prompt string) (string, error) {
	apiKey := a.config.APIKey
	if apiKey == "" && a.config.Provider == "gemini" {
		apiKey = a.config.GeminiAPIToken
	}
	if apiKey == "" || apiKey == "YOUR_GEMINI_API_TOKEN_HERE" {
		return "", fmt.Errorf("API key not configured: run 'agentai config set api_key <key>' or set in .agentai/config.json (--local or --global)")
	}
	a.ensureRateLimit()
	maxRetries := a.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		a.lastAt = time.Now()
		text, err := GenerateContent(ctx, a.config.Provider, apiKey, a.model, a.config.BaseURL, prompt)
		if err != nil {
			lastErr = err
			if isRetryableError(err) && attempt < maxRetries-1 {
				delay := extractRetryDelay(err.Error())
				a.sleep(delay)
				continue
			}
			return "", err
		}
		return text, nil
	}
	return "", lastErr
}

// GeneratePlan returns AI-generated plan text for the given prompt.
func (a *AICore) GeneratePlan(ctx context.Context, prompt string) (string, error) {
	text, err := a.generateContent(ctx, prompt)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			oldModel := a.model
			a.model = "gemini-2.5-flash"
			defer func() { a.model = oldModel }()
			return a.generateContent(ctx, prompt)
		}
		return "", fmt.Errorf("failed to generate plan: %w", err)
	}
	return text, nil
}

// GenerateCode returns AI-generated code for the given prompt.
func (a *AICore) GenerateCode(ctx context.Context, prompt string) (string, error) {
	text, err := a.generateContent(ctx, prompt)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			oldModel := a.model
			a.model = "gemini-2.5-flash"
			defer func() { a.model = oldModel }()
			return a.generateContent(ctx, prompt)
		}
		return "", fmt.Errorf("failed to generate code: %w", err)
	}
	return text, nil
}

// GenerateProjectName returns a short project name for the goal.
func (a *AICore) GenerateProjectName(ctx context.Context, goal string) (string, error) {
	prompt := fmt.Sprintf(`Generate a short, descriptive project name (2-4 words, lowercase, hyphenated) for this goal: "%s". 
Return ONLY the project name, nothing else. Examples: "todo-app", "express-api", "node-cli-tool"`, goal)
	text, err := a.generateContent(ctx, prompt)
	if err != nil {
		return SanitizeProjectName(goal), nil
	}
	return SanitizeProjectName(strings.TrimSpace(text)), nil
}

func SanitizeProjectName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.Trim(s, `'"`)
	for _, r := range []string{"  ", " ", "\t"} {
		s = strings.ReplaceAll(s, r, "-")
	}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			out = append(out, r)
		}
	}
	s = string(out)
	if len(s) > 50 {
		s = s[:50]
	}
	if s == "" {
		return "project"
	}
	return s
}

// ReasonAboutStep returns AI reasoning for executing the given step.
func (a *AICore) ReasonAboutStep(ctx context.Context, step *PlanStep, memory map[string]interface{}) (*types.Reasoning, error) {
	prompt := buildReasoningPrompt(step, memory)
	text, err := a.generateContent(ctx, prompt)
	if err != nil {
		return &types.Reasoning{
			Approach:       fmt.Sprintf("Execute %s for %s", step.Type, step.Description),
			Instructions:   fmt.Sprintf("Perform the %s operation as specified in the step", step.Type),
			Considerations: []string{"Follow best practices", "Maintain code quality"},
		}, nil
	}
	return parseReasoningResponse(text), nil
}

func buildReasoningPrompt(step *types.Step, memory map[string]interface{}) string {
	stepJSON, _ := json.MarshalIndent(step, "", "  ")
	memJSON, _ := json.MarshalIndent(memory, "", "  ")
	return fmt.Sprintf(`You are an AI code generator. Analyze the following step and provide reasoning.

Step: %s

Project Context:
%s

Provide:
1. Approach: How to execute this step
2. Instructions: Detailed instructions for execution
3. Considerations: Important factors to consider

Return as JSON:
{
  "approach": "...",
  "instructions": "...",
  "considerations": ["...", "..."]
}`, string(stepJSON), string(memJSON))
}

func parseReasoningResponse(response string) *types.Reasoning {
	re := regexp.MustCompile(`\{[\s\S]*\}`)
	if m := re.FindString(response); m != "" {
		var r types.Reasoning
		if json.Unmarshal([]byte(m), &r) == nil {
			return &r
		}
	}
	return &types.Reasoning{
		Approach:       "Standard execution",
		Instructions:   "Execute the step as specified",
		Considerations: []string{},
	}
}
