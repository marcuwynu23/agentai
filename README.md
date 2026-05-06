<h1 align="center">AgentAI</h1>
<p align="center">
  <a href="https://github.com/marcuwynu23/agentai/releases">
    <img src="https://img.shields.io/github/release/marcuwynu23/agentai.svg" alt="Latest Release">
  </a>
  <a href="https://goreportcard.com/report/github.com/marcuwynu23/agentai">
    <img src="https://goreportcard.com/badge/github.com/marcuwynu23/agentai" alt="Go Report Card">
  </a>
  <a href="https://godoc.org/github.com/marcuwynu23/agentai">
    <img src="https://img.shields.io/badge/godoc-reference-blue.svg" alt="GoDoc">
  </a>
  <a href="https://github.com/marcuwynu23/agentai/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/marcuwynu23/agentai.svg" alt="License">
  </a>
  <a href="https://github.com/marcuwynu23/agentai/stargazers">
    <img src="https://img.shields.io/github/stars/marcuwynu23/agentai.svg?style=social" alt="Stars">
  </a>
</p>


<p align="center">
  <strong>AI Code Assistant</strong>
</p>

<p align="center">
  A command-line tool that helps you write code faster with AI assistance, smart planning, and support for multiple AI providers.
</p>

<p align="center">
  <a href="#installation">Quick Start</a> •
  <a href="#features">Features</a> •
  <a href="#usage">Usage</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#providers">Providers</a>
</p>

## Features

### Goal-Based Development
- Describe what you want to build in plain English
- AI automatically creates a step-by-step plan
- Handles file creation, code writing, testing, and command execution

### Multiple AI Providers
- **Gemini** - Google's advanced AI model
- **OpenAI** - GPT models with strong capabilities
- **OpenRouter** - Access to many AI models through one service
- **Ollama** - Run AI models locally on your machine
- **Cloudflare AI Gateway** - Enterprise AI with analytics and caching

### Easy Configuration
- **Project Settings**: Store settings in your project folder
- **Global Settings**: Store settings for all projects
- **Environment Variables**: Override settings when needed
- **Switch Easily**: Change AI providers and models anytime

### Smart Memory
- **Project Memory**: Remembers your conversation and project state
- **Code Understanding**: Analyzes your existing code automatically
- **Context Awareness**: Keeps track of what you're working on

### Safe and Secure
- **Command Safety**: Validates shell commands before running
- **Local Processing**: Your code stays on your computer
- **No External Servers**: Built-in file operations

### Interactive Interface
- **Terminal UI**: Clean, modern interface
- **Real-time Updates**: See progress as it happens
- **Activity Logging**: Track what AgentAI is doing

## Installation

1. Go 1.22+

2. Clone or navigate to the project
   ```bash
   git clone https://github.com/marcuwynu23/agentai.git
   cd agentai
   ```

3. Build
   ```bash
   go build -o agentai .
   # Windows: agentai.exe
   ```

4. Configure (e.g., Gemini)
   ```bash
   agentai config set provider gemini --local
   agentai config set api_key YOUR_GEMINI_API_TOKEN --local
   agentai config set model gemini-2.5-flash --local
   ```
   Or use `.env` (see `.env.example`) with `GEMINI_API_TOKEN`, etc.

## Configuration

- Config file: `.agentai/config.json`
  - Local: `<current-directory>/.agentai/config.json` (use `--local`)
  - Global: `~/.agentai/config.json` (use `--global`)
- Keys: `provider`, `api_key`, `model`, `base_url`
- Resolution: Explicit `--config` path → local file → global file → environment variables

Commands:
```bash
agentai config show --local     # Show repo config
agentai config show --global    # Show user config
agentai config set provider ollama --local
agentai config set model llama3.2 --local
agentai config set base_url http://192.168.1.55:11434 --local  # Optional
```

Environment (optional):
- GEMINI_API_TOKEN – Gemini API key (used when provider is gemini and no `api_key` in file)
- OPENAI_API_KEY – OpenAI key (for openai provider)
- GEMINI_MODEL – Model name (e.g., `gemini-2.5-flash`)
- REQUEST_DELAY, MAX_RETRIES – Rate limiting
- WORKSPACE_PATH, LOGS_PATH – Paths (defaults: cwd, empty)

## Supported Providers

| Provider    | Default base URL                                          |
|-------------|------------------------------------------------------------|
| gemini      | `https://generativelanguage.googleapis.com/v1beta/models` |
| openai      | `https://api.openai.com/v1`                               |
| openrouter  | `https://openrouter.ai/api/v1`                            |
| ollama      | `http://localhost:11434`                                  |
| cloudflare  | `https://gateway.ai.cloudflare.com/v1`                    |

Override any with `base_url` in config.

#### Cloudflare AI Gateway Setup

For Cloudflare AI Gateway, you need:
1. **Account ID** - Find it in the Cloudflare dashboard
2. **API Token** - Create a token with AI Gateway - Read and AI Gateway - Edit permissions
3. **Base URL format** - `https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/compat`

Example configuration:
```bash
agentai config set provider cloudflare --local
agentai config set api_key YOUR_CLOUDFLARE_API_TOKEN --local
agentai config set base_url https://gateway.ai.cloudflare.com/v1/123456789/default/compat --local
agentai config set model openai/gpt-4 --local
```

## Usage

Run the interactive TUI chat:
```bash
agentai chat
```

Then type your goal and press enter to start!

With Ollama (default localhost):
```bash
agentai config set provider ollama --local
agentai config set model llama3.2 --local
agentai chat
```

## Development Status

- Multi-provider AI: Implemented (Gemini, OpenAI, OpenRouter, Ollama, Cloudflare AI Gateway) via raw HTTP
- Config: Local/global `.agentai/config.json` and env
- File operations: Create, modify, read in project directory
- Code generation: AI-generated code with cleanup
- Test creation: AI-generated or template test files
- Command execution: Validated, safe command execution
- TUI chat interface: Implemented with live activity logging

## How It Works

1. Config: Load provider, api_key, model, base_url (file + env).
2. Project: New run → AI suggests project name and directory; existing run → reuse.
3. Analysis: Scan codebase; summarize for planner.
4. Plan: AI produces a JSON plan (steps with types and dependencies).
5. Execution: For each step, AI reasons then the MCP client runs file/command/test logic.
6. Memory: Results and conversation saved to `<project-name>/.memory.json`.

## Requirements

- Go 1.22+
- API key for Gemini, OpenAI, OpenRouter, or Cloudflare AI Gateway; or local Ollama (no key)

## Documentation

### Core Documentation
- **[USAGE.md](USAGE.md)** – Complete usage guide, providers, configuration, and troubleshooting
- **[CONTRIBUTING.md](CONTRIBUTING.md)** – Development guidelines and contribution process
- **[CHANGELOG.md](CHANGELOG.md)** – Version history and release notes
- **[LICENSE](LICENSE)** – MIT license information

### Development Resources
- **[docs/architecture.md](docs/ARCHITECTURE.md)** – Technical architecture and design decisions

### Community Resources
- **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** – Community guidelines and code of conduct
- **[SECURITY.md](SECURITY.md)** – Security policies and vulnerability reporting
- **[RELEASE-NOTES.md](RELEASE-NOTES.md)** – Detailed release information

### Issue Templates
- **[.github/ISSUE_TEMPLATE/bug_report.md](.github/ISSUE_TEMPLATE/bug_report.md)** – Bug report template
- **[.github/ISSUE_TEMPLATE/feature_request.md](.github/ISSUE_TEMPLATE/feature_request.md)** – Feature request template
- **[.github/PULL_REQUEST_TEMPLATE.md](.github/PULL_REQUEST_TEMPLATE.md)** – Pull request template
