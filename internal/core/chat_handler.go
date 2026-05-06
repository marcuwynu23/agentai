package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agentai-go/internal/config"
	"agentai-go/internal/mcp"
)

// ChatCommand runs the full chat flow: memory, plan, execute.
func ChatCommand(ctx context.Context, goal string, cfg *config.Config) error {
	basePath, _ := os.Getwd()
	logsPath := cfg.LogsPath
	memoryManager := NewMemoryManager(basePath, logsPath)

	memory, err := memoryManager.LoadMemory()
	if err != nil {
		return err
	}
	conversationContext, _ := memoryManager.GetConversationContext()
	_ = memoryManager.AddConversationMessage("user", goal)

	planner := NewPlanner(cfg)
	aiCore := NewAICore(cfg)

	// Header
	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│   🤖 Agentic AI Code Assistant       │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()

	var projectName string
	var workspacePath string
	if pn, ok := memory["projectName"].(string); ok && pn != "" {
		projectName = pn
		workspacePath = filepath.Join(basePath, projectName)
		memoryManager = NewMemoryManager(workspacePath, logsPath)
		memory, _ = memoryManager.LoadMemory()
		conversationContext, _ = memoryManager.GetConversationContext()
		if _, err := os.Stat(workspacePath); err == nil {
			_ = os.Chdir(workspacePath)
		}
		fmt.Println("╭─ Project ───────────────────────────╮")
		fmt.Printf("│ Project: %-28s │\n", projectName)
		if conversationContext != "" && conversationContext != "No previous conversation." {
			fmt.Println("│ Continuing previous conversation...  │")
		}
		fmt.Println("╰─────────────────────────────────────╯")
	} else {
		fmt.Println("Generating project name...")
		projectName, err = aiCore.GenerateProjectName(ctx, goal)
		if err != nil {
			projectName = SanitizeProjectName(goal)
		}
		workspacePath = filepath.Join(basePath, projectName)
		_ = os.MkdirAll(workspacePath, 0755)
		_ = os.Chdir(workspacePath)
		memoryManager = NewMemoryManager(workspacePath, logsPath)
		_ = memoryManager.SetProjectName(projectName)
		memory, _ = memoryManager.LoadMemory()
		_ = memoryManager.AddConversationMessage("user", goal)
		fmt.Println("✓ Project name:", projectName)
		fmt.Println("╭─ Project ───────────────────────────╮")
		fmt.Printf("│ Project: %-28s │\n", projectName)
		fmt.Printf("│ Location: %-26s │\n", truncate(workspacePath, 26))
		fmt.Println("╰─────────────────────────────────────╯")
	}

	// Analyze codebase
	analyzer := NewCodebaseAnalyzer(workspacePath)
	analysis, err := analyzer.Analyze()
	if err != nil {
		analysis = &AnalysisResult{Summary: AnalysisSummary{}}
	}
	if analysis.Summary.TotalFiles > 0 {
		fmt.Printf("✓ Found %d files in codebase\n", analysis.Summary.TotalFiles)
		if analysis.Summary.TotalIssues > 0 {
			fmt.Printf("  ⚠ Detected %d issues (%d critical)\n", analysis.Summary.TotalIssues, analysis.Summary.CriticalIssues)
		}
	} else {
		fmt.Println("No existing codebase (new project)")
	}

	// Create plan
	fmt.Println("Generating plan...")
	plan, err := planner.CreatePlan(ctx, goal, memory, conversationContext, workspacePath)
	if err != nil {
		return fmt.Errorf("create plan: %w", err)
	}
	fmt.Printf("✓ Plan created with %d steps\n", len(plan.Steps))

	// Display plan
	fmt.Println()
	fmt.Println("╭─ Execution Plan ──────────────────────╮")
	for i, step := range plan.Steps {
		deps := ""
		if len(step.Dependencies) > 0 {
			deps = " (depends on: " + strings.Join(step.Dependencies, ", ") + ")"
		}
		fmt.Printf("│ %d. %s: %s%s\n", i+1, step.Type, truncate(step.Description, 35), deps)
		if step.Target != "" {
			fmt.Printf("│    → %s\n", step.Target)
		}
	}
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()

	_ = memoryManager.AddConversationMessage("assistant", fmt.Sprintf("Created plan with %d steps for: %s", len(plan.Steps), goal))
	memoryManager.LogAction("planning", map[string]interface{}{"goal": goal, "plan": plan})

	// Execute steps (inject code gen to avoid mcp->core import)
	genCode := func(ctx context.Context, prompt string) (string, error) { return aiCore.GenerateCode(ctx, prompt) }
	mcpClient := mcp.NewClient(workspacePath, genCode)
	fmt.Println("Executing plan...")
	fmt.Println()

	for i, step := range plan.Steps {
		fmt.Printf("[%d/%d] %s: %s\n", i+1, len(plan.Steps), strings.ToUpper(step.Type), step.Description)
		reasoning, _ := aiCore.ReasonAboutStep(ctx, step, memory)
		var result mcp.StepResult
		switch step.Type {
		case "file_creation":
			result = mcpClient.HandleFileCreation(ctx, step, reasoning)
		case "code_generation":
			result = mcpClient.HandleCodeGeneration(ctx, step, reasoning)
		case "test_creation":
			result = mcpClient.HandleTestCreation(ctx, step, reasoning)
		case "command_execution":
			result = mcpClient.HandleCommandExecution(step)
		default:
			result = mcp.StepResult{Success: false, Error: "unknown step type: " + step.Type}
		}
		_ = memoryManager.UpdateMemory(map[string]interface{}{
			"step": step.ID, "type": step.Type, "result": result, "timestamp": "",
		})
		if result.Success {
			fmt.Printf("  ✓ %s\n", result.Message)
		} else {
			fmt.Printf("  ✗ %s\n", result.Error)
		}
	}

	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│   ✅ Plan execution completed!       │")
	fmt.Println("╰─────────────────────────────────────╯")
	return memoryManager.SaveMemory(nil)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
