package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Default base URLs (overridable via config base_url).
const (
	defaultGeminiBase       = "https://generativelanguage.googleapis.com/v1beta/models"
	defaultOpenAIBase       = "https://api.openai.com/v1"
	defaultOpenRouterBase   = "https://openrouter.ai/api/v1"
	defaultOllamaBase       = "http://localhost:11434"
	defaultCloudflareBase   = "https://gateway.ai.cloudflare.com/v1"
)

// GenerateContent calls the configured provider (gemini, openai, openrouter, ollama).
// baseURL overrides the default for each provider when set.
func GenerateContent(ctx context.Context, provider, apiKey, model, baseURL, prompt string) (string, error) {
	switch strings.ToLower(provider) {
	case "gemini":
		url := baseURL
		if url == "" {
			url = defaultGeminiBase
		}
		return callGeminiGenerateContent(ctx, apiKey, model, url, prompt)
	case "openai":
		url := baseURL
		if url == "" {
			url = defaultOpenAIBase
		}
		return callOpenAICompletions(ctx, apiKey, model, url, prompt)
	case "openrouter":
		url := baseURL
		if url == "" {
			url = defaultOpenRouterBase
		}
		return callOpenAICompletions(ctx, apiKey, model, url, prompt)
	case "ollama":
		url := baseURL
		if url == "" {
			url = defaultOllamaBase
		}
		return callOllamaChat(ctx, model, url, prompt)
	case "cloudflare":
		url := baseURL
		if url == "" {
			url = defaultCloudflareBase
		}
		return callCloudflareAI(ctx, apiKey, model, url, prompt)
	default:
		return "", fmt.Errorf("unknown provider: %s (use: gemini, openai, openrouter, ollama, cloudflare)", provider)
	}
}

// OpenAI-compatible request/response (OpenAI + OpenRouter).
type openAIRequest struct {
	Model    string             `json:"model"`
	Messages []openAIMessage    `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func callOpenAICompletions(ctx context.Context, apiKey, model, baseURL, prompt string) (string, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := baseURL + "/chat/completions"
	body := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var out openAIResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("api error: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

// Ollama /api/chat (e.g. http://192.168.1.55:11434/api/chat)
type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaChatResponse struct {
	Message struct {
		Content string `json:"content"`
		Role    string `json:"role"`
	} `json:"message"`
	Error string `json:"error"`
}

func callOllamaChat(ctx context.Context, model, baseURL, prompt string) (string, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := baseURL + "/api/chat"
	body := ollamaChatRequest{
		Model:    model,
		Messages: []openAIMessage{{Role: "user", Content: prompt}},
		Stream:   false,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var out ollamaChatResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if out.Error != "" {
		return "", fmt.Errorf("ollama error: %s", out.Error)
	}
	return strings.TrimSpace(out.Message.Content), nil
}

// Cloudflare AI Gateway uses OpenAI-compatible endpoint but requires special headers
func callCloudflareAI(ctx context.Context, apiKey, model, baseURL, prompt string) (string, error) {
	// baseURL should include account_id and gateway_id: https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/compat
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := baseURL + "/chat/completions"
	
	body := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}
	
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	// Cloudflare AI Gateway uses cf-aig-authorization header instead of Authorization
	if apiKey != "" {
		req.Header.Set("cf-aig-authorization", "Bearer "+apiKey)
	}
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	var out openAIResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	
	if out.Error != nil {
		return "", fmt.Errorf("cloudflare ai gateway error: %s", out.Error.Message)
	}
	
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
