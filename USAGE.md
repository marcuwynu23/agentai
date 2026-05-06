# Usage Guide

A comprehensive guide to using AgentAI, an AI-powered code assistant with a modern Terminal User Interface (TUI).

## Getting Started

### 1. Build the Application

```bash
go build -o agentai .
# Windows: agentai.exe
# Linux/macOS: ./agentai
```

### 2. Configure AI Provider

Config is stored in `.agentai/config.json`. Use local (current directory) or global (home directory).

#### Option A: Configuration File (Recommended)

```bash
# Local: ./.agentai/config.json (per repository)
agentai config set provider gemini --local
agentai config set api_key YOUR_GEMINI_API_TOKEN --local
agentai config set model gemini-2.5-flash --local

# Global: ~/.agentai/config.json (all projects)
agentai config set provider gemini --global
agentai config set api_key YOUR_GEMINI_API_TOKEN --global
```

#### Option B: Environment Variables

Create a `.env` file (see `.env.example`) or export:

```bash
export GEMINI_API_TOKEN=your-actual-api-token-here
# Optional
export GEMINI_MODEL=gemini-2.5-flash
```

#### View Current Configuration

```bash
agentai config show --local    # repo config
agentai config show --global   # user config
```

### 3. Launch AgentAI TUI

```bash
# After building, launch the interactive TUI
./agentai chat

# Windows
agentai chat
```

The TUI provides an interactive chat interface where you can:
- Type your development goals in natural language
- See real-time progress as AgentAI plans and executes
- View formatted output with progress indicators
- Navigate through conversation history
- Exit cleanly with Ctrl+C

#### Example Goals for TUI

Once in the TUI, you can type goals like:

```
create a hello world program
create a REST API with Express.js that has a /users endpoint
build a todo list application with Node.js
```

---

## Supported AI Providers

AgentAI supports multiple AI backends using direct HTTP requests (no SDK dependencies). Configure the provider and optionally model and base_url settings.

| Provider       | Default Base URL                                              | Configuration Keys              |
|----------------|---------------------------------------------------------------|-------------------------------|
| gemini         | https://generativelanguage.googleapis.com/v1beta/models       | api_key, model, base_url        |
| openai         | https://api.openai.com/v1                                     | api_key, model, base_url        |
| openrouter      | https://openrouter.ai/api/v1                                  | api_key, model, base_url        |
| ollama          | http://localhost:11434                                        | model, base_url (no key)        |
| cloudflare      | https://gateway.ai.cloudflare.com/v1                          | api_key, model, base_url        |

The default URLs above are used automatically. Override with the base_url setting for custom endpoints (e.g., remote Ollama host or API proxy).

### Gemini

```bash
agentai config set provider gemini --local
agentai config set api_key YOUR_GEMINI_KEY --local
agentai config set model gemini-2.5-flash --local
# Optional: custom endpoint
agentai config set base_url https://your-proxy.com/v1beta/models --local
```

### OpenAI

```bash
agentai config set provider openai --local
agentai config set api_key sk-your-openai-key --local
agentai config set model gpt-4o-mini --local
```

### OpenRouter

```bash
agentai config set provider openrouter --local
agentai config set api_key sk-or-your-key --local
agentai config set model google/gemini-2.0-flash-001 --local
```

### Ollama (default: localhost)

```bash
agentai config set provider ollama --local
agentai config set model llama3.2 --local
# Optional: remote Ollama (default is http://localhost:11434)
agentai config set base_url http://192.168.1.55:11434 --local
```

---

## Configuration File Locations

- **`--local`**: `<current-directory>/.agentai/config.json` (repository-scoped)
- **`--global`**: `~/.agentai/config.json` (user-scoped)

If neither `--local` nor `--global` is specified, **show** and **set** commands use the local path. When running **chat**, configuration is resolved in this order: explicit `--config` path → local `.agentai/config.json` → global `~/.agentai/config.json` → environment variables.

---

## Understanding the TUI Interface

When you launch `agentai chat`, the TUI displays:

1. **Welcome Banner** – AgentAI logo and interface initialization
2. **Project Setup** – New projects receive AI-generated names; existing projects are detected
3. **Codebase Analysis** – File count and structure analysis displayed in real-time
4. **Planning Phase** – Step-by-step plan shown with dependencies
5. **Execution Display** – Live progress updates for each step
6. **Interactive Chat** – Continuous conversation interface with history

TUI Session Example:

```
Processing goal: create a simple calculator

+---------------------------------------+
|         AgentAI Code Assistant        |
+---------------------------------------+

Project name: simple-calculator
+---------------------------------------+
|           Execution Plan              |
+---------------------------------------+
| 1. file_creation: Create package.json     |
| 2. code_generation: Create calculator.js  |
| 3. test_creation: Add tests             |
| 4. command_execution: npm install        |
+---------------------------------------+

[1/4] FILE_CREATION: Create package.json
  File created: package.json
...

+---------------------------------------+
|       Plan execution completed!         |
+---------------------------------------+
```

---

## Generated Project Structure

- **`<project-name>/`** – Generated project directory (e.g., `simple-calculator/`)
- **`<project-name>/.memory.json`** – Project state and conversation history
- **`logs/`** – Log files (only created if `LOGS_PATH` environment variable is set)

---

## Command Reference

| Command | Description |
|---------|-------------|
| `agentai chat` | Launch interactive TUI for AI-assisted development |
| `agentai config show [--local\|--global]` | Show current AI configuration |
| `agentai config set <key> <value> [--local\|--global]` | Set `provider`, `api_key`, `model`, or `base_url` |
| `agentai version` | Display version, commit, and build information |

**Global Flags**

- `--config <path>` – Use a specific configuration file
- `-v, --verbose` – Enable verbose logging output

---

## Troubleshooting Guide

### Missing API Key

Configure an API key using either method:

```bash
agentai config set api_key YOUR_KEY --local
# or: export GEMINI_API_TOKEN=... or OPENAI_API_KEY=...
```

### Ollama Connection Issues

Start Ollama locally (`ollama serve`) or configure the remote host:

```bash
agentai config set base_url http://192.168.1.55:11434 --local
```

Default: `http://localhost:11434`

### Incorrect Provider or Model

Check and update your configuration:

```bash
agentai config show --local
agentai config set provider gemini --local
agentai config set model gemini-2.5-flash --local
```

### Go Version Requirements

AgentAI requires Go 1.22 or later. Verify your version with `go version`.

---

## Complete TUI Session Example

```bash
$ agentai config set provider ollama --local
$ agentai config set model llama3.2 --local
$ agentai chat

[AgentAI TUI launches with welcome screen]

> create a simple calculator

[Processing in real-time...]

+---------------------------------------+
|         AGENTAI CODE ASSISTANT       |
+---------------------------------------+

Analyzing goal: create a simple calculator
Generating project structure...
Project: simple-calculator

+---------------------------------------+
|           EXECUTION PLAN              |
+---------------------------------------+
| ✓ [1/4] Created package.json         |
| ✓ [2/4] Generated calculator.js        |
| ✓ [3/4] Created tests                |
| ✓ [4/4] Installed dependencies        |
+---------------------------------------+

+---------------------------------------+
|       PLAN EXECUTION COMPLETED!      |
+---------------------------------------+

[Continue typing next goal or type 'exit' to quit]
```
