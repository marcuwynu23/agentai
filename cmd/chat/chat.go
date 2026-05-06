package chat

import (
	"context"
	"fmt"
	"os"
	"strings"

	"agentai-go/internal/config"
	"agentai-go/internal/core"
	"agentai-go/internal/mcp"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	bannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00d4ff")).
			Bold(true).
			Align(lipgloss.Center)

	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			PaddingLeft(2)

	assistantMsgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4ade80")).
				PaddingLeft(2)

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			PaddingLeft(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f87171")).
			PaddingLeft(2)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00d4ff"))
)

const banner = `
 █████   ██████  ███████ ███    ██ ████████  █████  ██ 
██   ██ ██       ██      ████   ██    ██    ██   ██ ██ 
███████ ██   ███ █████   ██ ██  ██    ██    ███████ ██ 
██   ██ ██    ██ ██      ██  ██ ██    ██    ██   ██ ██ 
██   ██  ██████  ███████ ██   ████    ██    ██   ██ ██ 
`

type message struct {
	role    string
	content string
}

type model struct {
	spinner       spinner.Model
	textInput     textinput.Model
	messages      []message
	logs          []string
	viewport      viewport.Model
	loading       bool
	err           error
	ctx           context.Context
	cfg           *config.Config
	logChan       chan string
	responseChan  chan responseMsg
	width         int
	height        int
	ready         bool
}

func initialModel(ctx context.Context, cfg *config.Config, logChan chan string, responseChan chan responseMsg) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d4ff"))

	ti := textinput.New()
	ti.Placeholder = "Type your message here... (Ctrl+C to quit)"
	ti.Focus()
	ti.CharLimit = 500

	vp := viewport.New(80, 10)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)

	return model{
		spinner:       s,
		textInput:     ti,
		messages:      []message{},
		logs:          []string{},
		viewport:      vp,
		loading:       false,
		ctx:           ctx,
		cfg:           cfg,
		logChan:       logChan,
		responseChan:  responseChan,
		width:         80,
		height:        24,
		ready:         false,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.waitForMessages)
}

func (m model) waitForMessages() tea.Msg {
	select {
	case log := <-m.logChan:
		return logMsg(log)
	case resp := <-m.responseChan:
		return resp
	case <-m.ctx.Done():
		return nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.textInput.Value() != "" && !m.loading {
				userMsg := m.textInput.Value()
				m.messages = append(m.messages, message{role: "user", content: userMsg})
				m.textInput.SetValue("")
				m.logs = append(m.logs, "")
				m.loading = true

				// Start processing in a goroutine so log channel doesn't block
				go func() {
					m.responseChan <- m.processMessage(userMsg)
				}()

				return m, tea.Batch(m.spinner.Tick, m.waitForMessages)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = max(10, msg.Height-35)
		m.textInput.Width = max(40, msg.Width-8)
		if !m.ready {
			m.ready = true
		}
	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, tea.Batch(cmd, m.waitForMessages)
		}
	case responseMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.messages = append(m.messages, message{role: "assistant", content: msg.content})
		}
		return m, m.waitForMessages
	case logMsg:
		m.logs = append(m.logs, string(msg))
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		return m, m.waitForMessages
	}

	m.textInput, cmd = m.textInput.Update(msg)
	m.viewport, _ = m.viewport.Update(msg)
	return m, cmd
}

type responseMsg struct {
	content string
	err     error
}

type logMsg string

func (m model) processMessage(userMsg string) responseMsg {
	apiKey := m.cfg.APIKey
	if apiKey == "" {
		apiKey = m.cfg.GeminiAPIToken
	}
	if apiKey == "" || apiKey == "YOUR_GEMINI_API_TOKEN_HERE" {
		return responseMsg{err: fmt.Errorf("no API key. Set with 'agentai config set api_key <key>' (--local or --global) or .env (API_KEY)")}
	}

	goal := userMsg

	basePath, _ := os.Getwd()
	logsPath := m.cfg.LogsPath
	memoryManager := core.NewMemoryManager(basePath, logsPath)

	memory, err := memoryManager.LoadMemory()
	if err != nil {
		return responseMsg{err: err}
	}
	conversationContext, _ := memoryManager.GetConversationContext()
	_ = memoryManager.AddConversationMessage("user", goal)

	planner := core.NewPlanner(m.cfg)
	aiCore := core.NewAICore(m.cfg)

	var projectName string
	var workspacePath string
	if pn, ok := memory["projectName"].(string); ok && pn != "" {
		projectName = pn
		workspacePath = fmt.Sprintf("%s/%s", basePath, projectName)
		memoryManager = core.NewMemoryManager(workspacePath, logsPath)
		memory, _ = memoryManager.LoadMemory()
		conversationContext, _ = memoryManager.GetConversationContext()
		if _, err := os.Stat(workspacePath); err == nil {
			_ = os.Chdir(workspacePath)
		}
		m.logChan <- "╭─ Project ───────────────────────────╮"
		m.logChan <- fmt.Sprintf("│ Project: %-28s │", projectName)
		if conversationContext != "" && conversationContext != "No previous conversation." {
			m.logChan <- "│ Continuing previous conversation...  │"
		}
		m.logChan <- "╰─────────────────────────────────────╯"
	} else {
		m.logChan <- "Generating project name..."
		projectName, err = aiCore.GenerateProjectName(m.ctx, goal)
		if err != nil {
			projectName = core.SanitizeProjectName(goal)
		}
		workspacePath = fmt.Sprintf("%s/%s", basePath, projectName)
		_ = os.MkdirAll(workspacePath, 0755)
		_ = os.Chdir(workspacePath)
		memoryManager = core.NewMemoryManager(workspacePath, logsPath)
		_ = memoryManager.SetProjectName(projectName)
		memory, _ = memoryManager.LoadMemory()
		_ = memoryManager.AddConversationMessage("user", goal)
		m.logChan <- fmt.Sprintf("✓ Project name: %s", projectName)
		m.logChan <- "╭─ Project ───────────────────────────╮"
		m.logChan <- fmt.Sprintf("│ Project: %-28s │", projectName)
		m.logChan <- fmt.Sprintf("│ Location: %-26s │", truncate(workspacePath, 26))
		m.logChan <- "╰─────────────────────────────────────╯"
	}

	analyzer := core.NewCodebaseAnalyzer(workspacePath)
	analysis, err := analyzer.Analyze()
	if err != nil {
		analysis = &core.AnalysisResult{Summary: core.AnalysisSummary{}}
	}
	if analysis.Summary.TotalFiles > 0 {
		m.logChan <- fmt.Sprintf("✓ Found %d files in codebase", analysis.Summary.TotalFiles)
		if analysis.Summary.TotalIssues > 0 {
			m.logChan <- fmt.Sprintf("  ⚠ Detected %d issues (%d critical)", analysis.Summary.TotalIssues, analysis.Summary.CriticalIssues)
		}
	} else {
		m.logChan <- "No existing codebase (new project)"
	}

	m.logChan <- "Generating plan..."
	plan, err := planner.CreatePlan(m.ctx, goal, memory, conversationContext, workspacePath)
	if err != nil {
		return responseMsg{err: fmt.Errorf("create plan: %w", err)}
	}
	m.logChan <- fmt.Sprintf("✓ Plan created with %d steps", len(plan.Steps))

	m.logChan <- ""
	m.logChan <- "╭─ Execution Plan ──────────────────────╮"
	for i, step := range plan.Steps {
		deps := ""
		if len(step.Dependencies) > 0 {
			deps = " (depends on: " + strings.Join(step.Dependencies, ", ") + ")"
		}
		m.logChan <- fmt.Sprintf("│ %d. %s: %s%s", i+1, step.Type, truncate(step.Description, 35), deps)
		if step.Target != "" {
			m.logChan <- fmt.Sprintf("│    → %s", step.Target)
		}
	}
	m.logChan <- "╰─────────────────────────────────────╯"
	m.logChan <- ""

	_ = memoryManager.AddConversationMessage("assistant", fmt.Sprintf("Created plan with %d steps for: %s", len(plan.Steps), goal))
	memoryManager.LogAction("planning", map[string]interface{}{"goal": goal, "plan": plan})

	genCode := func(ctx context.Context, prompt string) (string, error) { return aiCore.GenerateCode(ctx, prompt) }
	mcpClient := mcp.NewClient(workspacePath, genCode)
	m.logChan <- "Executing plan..."
	m.logChan <- ""

	results := []string{}
	for i, step := range plan.Steps {
		m.logChan <- fmt.Sprintf("[%d/%d] %s: %s", i+1, len(plan.Steps), strings.ToUpper(step.Type), step.Description)
		reasoning, _ := aiCore.ReasonAboutStep(m.ctx, step, memory)
		var result mcp.StepResult
		switch step.Type {
		case "file_creation":
			result = mcpClient.HandleFileCreation(m.ctx, step, reasoning)
		case "code_generation":
			result = mcpClient.HandleCodeGeneration(m.ctx, step, reasoning)
		case "test_creation":
			result = mcpClient.HandleTestCreation(m.ctx, step, reasoning)
		case "command_execution":
			result = mcpClient.HandleCommandExecution(step)
		default:
			result = mcp.StepResult{Success: false, Error: "unknown step type: " + step.Type}
		}
		_ = memoryManager.UpdateMemory(map[string]interface{}{
			"step": step.ID, "type": step.Type, "result": result, "timestamp": "",
		})
		if result.Success {
			m.logChan <- fmt.Sprintf("  ✓ %s", result.Message)
			results = append(results, fmt.Sprintf("✓ Step %d/%d: %s", i+1, len(plan.Steps), result.Message))
		} else {
			m.logChan <- fmt.Sprintf("  ✗ %s", result.Error)
			results = append(results, fmt.Sprintf("✗ Step %d/%d: %s", i+1, len(plan.Steps), result.Error))
		}
	}

	m.logChan <- ""
	m.logChan <- "╭─────────────────────────────────────╮"
	m.logChan <- "│   ✅ Plan execution completed!       │"
	m.logChan <- "╰─────────────────────────────────────╯"
	_ = memoryManager.SaveMemory(nil)
	return responseMsg{content: strings.Join(results, "\n"), err: nil}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var s strings.Builder

	bannerRendered := bannerStyle.Width(m.width).Render(banner)
	s.WriteString(bannerRendered)
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("❌ Error: %v\n", m.err)))
		s.WriteString("\n")
	}

	for _, msg := range m.messages {
		if msg.role == "user" {
			s.WriteString(userMsgStyle.Render(fmt.Sprintf("> %s\n", msg.content)))
		} else {
			s.WriteString(assistantMsgStyle.Render(fmt.Sprintf("✦ %s\n", msg.content)))
		}
		s.WriteString("\n")
	}

	if len(m.logs) > 0 {
		s.WriteString(m.viewport.View())
		s.WriteString("\n")
	}

	if m.loading {
		s.WriteString(fmt.Sprintf("%s Processing...\n\n", m.spinner.View()))
	}

	inputRendered := borderStyle.Width(m.width - 4).Render(m.textInput.View())
	s.WriteString(inputRendered)
	s.WriteString("\n")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, s.String())
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Interactive chat interface",
		Long:  "Interactive terminal chat interface for code generation.",
		Args:  cobra.NoArgs,
		RunE:  runChat,
	}
	return cmd
}

func runChat(cmd *cobra.Command, args []string) error {
	_ = godotenv.Load() // .env in cwd
	configPath, _ := cmd.Flags().GetString("config")
	cwd, _ := os.Getwd()
	cfg := config.Load(configPath, cwd)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logChan := make(chan string, 100)
	responseChan := make(chan responseMsg, 1)
	p := tea.NewProgram(initialModel(ctx, cfg, logChan, responseChan), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
