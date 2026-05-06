# AgentAI Architecture

High-level layout of the AgentAI CLI with TUI interface and its packages.

---

## Layers

### CLI (`cmd/`)

- **root.go**: Root command `agentai`, persistent flags (`--config`, `-v`), registers `chat`, `config`, `version` subcommands.
- **chat/**: TUI-based chat interface using Bubbletea, loads config, validates API key, orchestrates TUI workflow.
- **config/**: `config show` and `config set` commands for `.agentai/config.json` (--local | --global).
- **version/**: Prints version, commit, build date (from root's build-time vars).

No business logic in `cmd/`; it delegates to `internal/`. The chat command implements a full TUI with real-time updates.

### Config (`internal/config`)

- **config.go**: `Load(explicitPath, cwd)` – merges file (local then global) and env (provider, api_key, model, base_url, etc.).
- **file.go**: Paths for local/global `.agentai/config.json`, `LoadAIConfig`, `SaveAIConfig`, `ResolveAIConfig`.

### Core (`internal/core`)

- **ai_core.go**: Rate limiting, retries, `GeneratePlan`, `GenerateCode`, `ReasonAboutStep`, `GenerateProjectName` (dispatches to providers).
- **gemini.go**: Gemini API HTTP client implementation.
- **providers.go**: Unified provider interface supporting Gemini, OpenAI, OpenRouter, Ollama, and Cloudflare AI Gateway via raw HTTP.
- **planner.go**: Builds planning prompts, calls AI, parses JSON plans into executable steps.
- **memory_manager.go**: Load/save `.memory.json`, manages conversation history and execution context.
- **codebase_analyzer.go**: Scans workspace, analyzes project structure, detects issues, formats analysis for AI consumption.
- **chat_handler.go**: Orchestrates complete chat workflow: memory management, project naming, codebase analysis, planning, and step execution via MCP client.
- **types.go**: Type definitions and aliases (Plan, Step, Reasoning) shared across packages.

### Types (`internal/types`)

- **types.go**: `Plan`, `Step`, `Reasoning` – shared between core and mcp to avoid import cycles.

### MCP (`internal/mcp`)

- **client.go**: Coordinates file/command/test “servers”; implements step handlers (file_creation, code_generation, test_creation, command_execution). Uses injected `CodeGenFunc` from core.
- **file_server.go**: Create, modify, read files under workspace.
- **command_server.go**: Validate and execute shell commands (allow/block lists).
- **test_server.go**: Create test files (AI or template).

---

## Flow (TUI Chat)

```
User: agentai chat
    ↓
cmd/chat: Initialize TUI (Bubbletea), load config, validate API key
    ↓
TUI Loop: User input → Process → Display results
    ↓
User enters goal: "build a todo app"
    ↓
core.ChatCommand: load memory, resolve project (new name or existing)
    ↓
CodebaseAnalyzer.Analyze() → TUI shows analysis progress
    ↓
Planner.CreatePlan(goal, memory, conversation, analysis) → TUI displays plan
    ↓
For each step (with TUI progress updates):
    AICore.ReasonAboutStep(step, memory) → TUI shows reasoning
    MCPClient.Handle*(step, reasoning) → FileServer / CommandServer / TestServer
    MemoryManager.UpdateMemory(...) → TUI logs activity
    ↓
MemoryManager.SaveMemory() → TUI shows completion
    ↓
TUI returns to input prompt for next goal
```

---

## AI Provider Architecture

### Provider Support
- **Gemini**: Google's generative AI API with custom HTTP client
- **OpenAI**: GPT models via OpenAI API endpoints
- **OpenRouter**: Multi-provider access through unified API
- **Ollama**: Local model hosting with HTTP interface
- **Cloudflare AI Gateway**: Enterprise AI with analytics and custom headers

### Provider Implementation Pattern
All providers follow the same pattern:
1. HTTP client with proper error handling
2. Request/response structures for each provider
3. Retry logic and rate limiting in AICore
4. Standardized interface through GenerateContent()

## TUI Architecture

### Bubbletea Components
- **Model**: Main TUI state management
- **Update**: Handles user input and async events
- **View**: Renders interface with styled components
- **Spinner**: Loading indicators during AI processing
- **TextInput**: User input field with validation
- **Viewport**: Scrollable message history

### Real-time Updates
- Progress indicators for plan execution
- Live logging of AI operations
- Interactive step-by-step execution display
- Conversation history with formatted messages

## Tests

The `test/` directory is reserved for AgentAI-specific tests (config, providers, chat, TUI).

Run: `go test ./...`

### Test Coverage Areas
- Provider implementations with mock responses
- Configuration parsing and validation
- TUI component behavior
- Memory management operations
- MCP server functionality
