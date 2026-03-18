// Package aiprovider provides AI provider compatibility layer.
// Handles request formatting, URL normalization, and error mapping.
package aiprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ProviderType represents the category of AI provider.
type ProviderType string

const (
	ProviderTypeOpenAI    ProviderType = "openai_compatible" // OpenAI-compatible API
	ProviderTypeAnthropic ProviderType = "anthropic"          // Anthropic Claude
	ProviderTypeGemini    ProviderType = "gemini"             // Google Gemini
	ProviderTypeOllama    ProviderType = "ollama"             // Ollama local
)

// ProviderConfig contains provider-specific configuration.
type ProviderConfig struct {
	Type            ProviderType
	AuthHeader      string            // Header name for API key
	AuthPrefix      string            // Prefix for auth value (e.g., "Bearer ")
	ExtraHeaders    map[string]string // Additional headers required
	RequestFormat   string            // "openai" or "anthropic" or "gemini"
	DefaultEndpoint string            // Default API endpoint
}

// Provider registry with their configurations.
var providerRegistry = map[string]ProviderConfig{
	// OpenAI-compatible providers (China)
	"deepseek": {
		Type:            ProviderTypeOpenAI,
		AuthHeader:      "Authorization",
		AuthPrefix:      "Bearer ",
		RequestFormat:   "openai",
		DefaultEndpoint: "/v1/chat/completions",
	},
	"qwen": {
		Type:            ProviderTypeOpenAI,
		AuthHeader:      "Authorization",
		AuthPrefix:      "Bearer ",
		RequestFormat:   "openai",
		DefaultEndpoint: "/v1/chat/completions",
	},
	"doubao": {
		Type:            ProviderTypeOpenAI,
		AuthHeader:      "Authorization",
		AuthPrefix:      "Bearer ",
		RequestFormat:   "openai",
		DefaultEndpoint: "/chat/completions",
	},
	"glm": {
		Type:            ProviderTypeOpenAI,
		AuthHeader:      "Authorization",
		AuthPrefix:      "Bearer ",
		RequestFormat:   "openai",
		DefaultEndpoint: "/v4/chat/completions",
	},
	"minimax": {
		Type:            ProviderTypeOpenAI,
		AuthHeader:      "Authorization",
		AuthPrefix:      "Bearer ",
		RequestFormat:   "openai",
		DefaultEndpoint: "/v1/chat/completions",
	},
	"moonshot": {
		Type:            ProviderTypeOpenAI,
		AuthHeader:      "Authorization",
		AuthPrefix:      "Bearer ",
		RequestFormat:   "openai",
		DefaultEndpoint: "/v1/chat/completions",
	},

	// OpenAI-compatible providers (International)
	"openai": {
		Type:            ProviderTypeOpenAI,
		AuthHeader:      "Authorization",
		AuthPrefix:      "Bearer ",
		RequestFormat:   "openai",
		DefaultEndpoint: "/v1/chat/completions",
	},
	"xai": {
		Type:            ProviderTypeOpenAI,
		AuthHeader:      "Authorization",
		AuthPrefix:      "Bearer ",
		RequestFormat:   "openai",
		DefaultEndpoint: "/v1/chat/completions",
	},

	// Special format providers
	"anthropic": {
		Type:            ProviderTypeAnthropic,
		AuthHeader:      "x-api-key",
		AuthPrefix:      "",
		RequestFormat:   "anthropic",
		DefaultEndpoint: "/v1/messages",
		ExtraHeaders: map[string]string{
			"anthropic-version": "2023-06-01",
		},
	},
	"gemini": {
		Type:            ProviderTypeGemini,
		AuthHeader:      "", // API key goes in URL param
		AuthPrefix:      "",
		RequestFormat:   "gemini",
		DefaultEndpoint: "/v1beta/models/{model}:generateContent",
	},

	// Local providers
	"ollama": {
		Type:            ProviderTypeOllama,
		AuthHeader:      "",
		AuthPrefix:      "",
		RequestFormat:   "ollama",
		DefaultEndpoint: "/api/chat",
	},
}

// NormalizeURL normalizes and validates the API URL.
// - Ensures protocol is present
// - Removes duplicate slashes
// - Safely joins host and endpoint
func NormalizeURL(host, endpoint string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", fmt.Errorf("API 主机不能为空")
	}

	// Add protocol if missing
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}

	// Parse and validate host
	parsedURL, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("无效的 API 主机: %w", err)
	}

	// Rebuild host without trailing slash
	host = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Normalize endpoint
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return host, nil
	}

	// Ensure endpoint starts with /
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	// Remove duplicate slashes in path
	fullURL := host + endpoint
	fullURL = strings.ReplaceAll(fullURL, "//", "/")
	// Fix protocol slashes
	fullURL = strings.ReplaceAll(fullURL, "http:/", "http://")
	fullURL = strings.ReplaceAll(fullURL, "https:/", "https://")

	return fullURL, nil
}

// GetProviderConfig returns the configuration for a provider.
// Returns default OpenAI-compatible config for unknown providers.
func GetProviderConfig(provider string) ProviderConfig {
	if cfg, ok := providerRegistry[provider]; ok {
		return cfg
	}
	// Default to OpenAI-compatible
	return ProviderConfig{
		Type:            ProviderTypeOpenAI,
		AuthHeader:      "Authorization",
		AuthPrefix:      "Bearer ",
		RequestFormat:   "openai",
		DefaultEndpoint: "/v1/chat/completions",
	}
}

// TestRequest contains parameters for AI API test.
type TestRequest struct {
	Provider    string
	APIHost     string
	APIEndpoint string
	APIKey      string
	Model       string
}

// TestResult contains the result of AI API test.
type TestResult struct {
	Success   bool
	LatencyMs int64
	Message   string
	Error     string
}

// TestConnection tests the AI API connection.
func TestConnection(ctx context.Context, req TestRequest) TestResult {
	startTime := time.Now()

	// Validate required fields
	if req.Provider == "" {
		return TestResult{Success: false, Error: "未选择 AI 提供商"}
	}
	if req.APIHost == "" {
		return TestResult{Success: false, Error: "API 主机不能为空"}
	}
	if req.APIKey == "" && req.Provider != "ollama" {
		return TestResult{Success: false, Error: "API 密钥不能为空"}
	}
	if req.Model == "" {
		return TestResult{Success: false, Error: "模型不能为空"}
	}

	// Get provider config
	cfg := GetProviderConfig(req.Provider)

	// Determine endpoint
	endpoint := req.APIEndpoint
	if endpoint == "" {
		endpoint = cfg.DefaultEndpoint
		// For Gemini, inject model name into endpoint
		if cfg.Type == ProviderTypeGemini {
			endpoint = strings.ReplaceAll(endpoint, "{model}", req.Model)
		}
	}

	// Normalize URL
	apiURL, err := NormalizeURL(req.APIHost, endpoint)
	if err != nil {
		return TestResult{Success: false, Error: err.Error()}
	}

	slog.Info("Testing AI connection",
		"provider", req.Provider,
		"url", apiURL,
		"model", req.Model)

	// Build request body based on provider type
	var bodyBytes []byte
	switch cfg.Type {
	case ProviderTypeAnthropic:
		bodyBytes, err = buildAnthropicRequest(req.Model)
	case ProviderTypeGemini:
		bodyBytes, err = buildGeminiRequest()
	case ProviderTypeOllama:
		bodyBytes, err = buildOllamaRequest(req.Model)
	default:
		bodyBytes, err = buildOpenAIRequest(req.Model)
	}

	if err != nil {
		return TestResult{Success: false, Error: fmt.Sprintf("构建请求失败: %v", err)}
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return TestResult{Success: false, Error: fmt.Sprintf("创建请求失败: %v", err)}
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Set auth header if required
	if cfg.AuthHeader != "" && req.APIKey != "" {
		authValue := cfg.AuthPrefix + req.APIKey
		httpReq.Header.Set(cfg.AuthHeader, authValue)
	}

	// Set extra headers
	for key, value := range cfg.ExtraHeaders {
		httpReq.Header.Set(key, value)
	}

	// For Gemini, API key goes in URL param instead
	if cfg.Type == ProviderTypeGemini {
		q := httpReq.URL.Query()
		q.Set("key", req.APIKey)
		httpReq.URL.RawQuery = q.Encode()
	}

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return mapNetworkError(err)
	}
	defer resp.Body.Close()

	latencyMs := time.Since(startTime).Milliseconds()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)

	// Check status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		slog.Info("AI connection test successful",
			"provider", req.Provider,
			"latency_ms", latencyMs)
		return TestResult{
			Success:   true,
			LatencyMs: latencyMs,
			Message:   fmt.Sprintf("连接成功 (%dms)", latencyMs),
		}
	}

	// Map error response
	errMsg := mapAPIError(resp.StatusCode, respBody, req.Provider)
	slog.Error("AI connection test failed",
		"provider", req.Provider,
		"status", resp.StatusCode,
		"error", errMsg)
	return TestResult{
		Success:   false,
		LatencyMs: latencyMs,
		Error:     errMsg,
	}
}

// Request builders for different provider types

func buildOpenAIRequest(model string) ([]byte, error) {
	body := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "Hi"},
		},
		"max_tokens": 5,
	}
	return json.Marshal(body)
}

func buildAnthropicRequest(model string) ([]byte, error) {
	body := map[string]interface{}{
		"model":      model,
		"max_tokens": 5,
		"messages": []map[string]string{
			{"role": "user", "content": "Hi"},
		},
	}
	return json.Marshal(body)
}

func buildGeminiRequest() ([]byte, error) {
	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": "Hi"}}},
		},
	}
	return json.Marshal(body)
}

func buildOllamaRequest(model string) ([]byte, error) {
	body := map[string]interface{}{
		"model":   model,
		"message": "Hi",
		"stream":  false,
	}
	return json.Marshal(body)
}

// Error mapping functions

func mapNetworkError(err error) TestResult {
	errStr := err.Error()

	// Timeout
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded") {
		return TestResult{Success: false, Error: "连接超时，请检查网络或主机地址"}
	}

	// Connection refused
	if strings.Contains(errStr, "connection refused") {
		return TestResult{Success: false, Error: "无法连接到服务器，请检查主机地址和端口"}
	}

	// No such host
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "lookup") {
		return TestResult{Success: false, Error: "无法解析主机名，请检查 API 主机地址"}
	}

	// TLS/SSL errors
	if strings.Contains(errStr, "certificate") || strings.Contains(errStr, "TLS") {
		return TestResult{Success: false, Error: "SSL/TLS 证书错误"}
	}

	// Network unreachable
	if strings.Contains(errStr, "network is unreachable") {
		return TestResult{Success: false, Error: "网络不可达"}
	}

	return TestResult{Success: false, Error: fmt.Sprintf("网络错误: %v", err)}
}

func mapAPIError(statusCode int, body []byte, provider string) string {
	// Try to extract error message from response
	var errResp struct {
		Error   string `json:"error"`
		Detail  string `json:"detail"`
		Message string `json:"message"`
	}
	json.Unmarshal(body, &errResp)

	errMsg := errResp.Error
	if errMsg == "" {
		errMsg = errResp.Detail
	}
	if errMsg == "" {
		errMsg = errResp.Message
	}

	switch statusCode {
	case 400:
		if errMsg != "" {
			return fmt.Sprintf("请求格式错误: %s", errMsg)
		}
		return "请求格式错误"
	case 401:
		return "API 密钥无效或已过期"
	case 403:
		return "访问被拒绝，请检查 API 密钥权限"
	case 404:
		return "API 端点不存在，请检查主机和端点地址"
	case 429:
		return "请求过于频繁，请稍后重试"
	case 500, 502, 503:
		return fmt.Sprintf("%s 服务暂时不可用，请稍后重试", provider)
	default:
		if errMsg != "" {
			return fmt.Sprintf("错误 (%d): %s", statusCode, errMsg)
		}
		return fmt.Sprintf("请求失败 (HTTP %d)", statusCode)
	}
}

// ============================================================
// Model Query Functions
// ============================================================

// ModelInfo represents a single model's information.
type ModelInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
	Owner  string `json:"owner,omitempty"`
	Object string `json:"object,omitempty"`
}

// QueryModelsResult contains the result of model query.
type QueryModelsResult struct {
	Success bool         `json:"success"`
	Models  []ModelInfo  `json:"models,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// QueryModelsRequest contains parameters for model list query.
type QueryModelsRequest struct {
	Provider string `json:"provider"`
	APIHost  string `json:"api_host"`
	APIKey   string `json:"api_key"`
}

// QueryModels queries available models from the AI provider.
// For cloud providers: uses /v1/models endpoint with API key authentication.
// For Ollama (local): uses /api/tags endpoint without authentication.
func QueryModels(ctx context.Context, req QueryModelsRequest) QueryModelsResult {
	slog.Info("QueryModels called", "provider", req.Provider, "api_host", req.APIHost)

	// Validate required fields
	if req.Provider == "" {
		return QueryModelsResult{Success: false, Error: "未选择 AI 提供商"}
	}
	if req.APIHost == "" {
		return QueryModelsResult{Success: false, Error: "API 主机不能为空"}
	}

	cfg := GetProviderConfig(req.Provider)

	switch cfg.Type {
	case ProviderTypeOllama:
		return queryOllamaModels(ctx, req.APIHost)
	default:
		// Cloud providers require API key
		if req.APIKey == "" {
			return QueryModelsResult{Success: false, Error: "API 密钥不能为空"}
		}
		return queryOpenAIModels(ctx, req.Provider, req.APIHost, req.APIKey, cfg)
	}
}

// queryOllamaModels queries models from local Ollama instance.
// Uses GET /api/tags endpoint which returns list of pulled models.
func queryOllamaModels(ctx context.Context, apiHost string) QueryModelsResult {
	// Normalize URL
	apiURL, err := NormalizeURL(apiHost, "/api/tags")
	if err != nil {
		return QueryModelsResult{Success: false, Error: err.Error()}
	}

	slog.Info("Querying Ollama models", "url", apiURL)

	// Create HTTP request (no auth needed for Ollama)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return QueryModelsResult{Success: false, Error: fmt.Sprintf("创建请求失败: %v", err)}
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		slog.Error("Ollama model query failed", "error", err)
		return QueryModelsResult{Success: false, Error: mapQueryNetworkError(err)}
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		errMsg := fmt.Sprintf("Ollama 返回错误 (HTTP %d)", resp.StatusCode)
		slog.Error("Ollama model query failed", "status", resp.StatusCode, "body", string(respBody))
		return QueryModelsResult{Success: false, Error: errMsg}
	}

	// Parse Ollama response format: {"models": [{"name": "llama2:latest", ...}]}
	var ollamaResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return QueryModelsResult{Success: false, Error: fmt.Sprintf("解析响应失败: %v", err)}
	}

	// Convert to ModelInfo slice
	models := make([]ModelInfo, 0, len(ollamaResp.Models))
	for _, m := range ollamaResp.Models {
		models = append(models, ModelInfo{
			ID:   m.Name,
			Name: m.Name,
		})
	}

	slog.Info("Ollama models queried successfully", "count", len(models))
	return QueryModelsResult{
		Success: true,
		Models:  models,
	}
}

// queryOpenAIModels queries models from OpenAI-compatible API.
// Uses GET /v1/models endpoint with Bearer token authentication.
func queryOpenAIModels(ctx context.Context, provider, apiHost, apiKey string, cfg ProviderConfig) QueryModelsResult {
	// Normalize URL
	apiURL, err := NormalizeURL(apiHost, "/v1/models")
	if err != nil {
		return QueryModelsResult{Success: false, Error: err.Error()}
	}

	slog.Info("Querying OpenAI-compatible models", "provider", provider, "url", apiURL)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return QueryModelsResult{Success: false, Error: fmt.Sprintf("创建请求失败: %v", err)}
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Set auth header
	if cfg.AuthHeader != "" {
		authValue := cfg.AuthPrefix + apiKey
		httpReq.Header.Set(cfg.AuthHeader, authValue)
	}

	// Execute request
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		slog.Error("OpenAI model query failed", "provider", provider, "error", err)
		return QueryModelsResult{Success: false, Error: mapQueryNetworkError(err)}
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		errMsg := mapAPIError(resp.StatusCode, respBody, provider)
		slog.Error("OpenAI model query failed", "provider", provider, "status", resp.StatusCode, "error", errMsg)
		return QueryModelsResult{Success: false, Error: errMsg}
	}

	// Parse OpenAI response format: {"data": [{"id": "gpt-4", "object": "model", ...}]}
	var openAIResp struct {
		Data []struct {
			ID     string `json:"id"`
			Object string `json:"object"`
			Owner  string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return QueryModelsResult{Success: false, Error: fmt.Sprintf("解析响应失败: %v", err)}
	}

	// Convert to ModelInfo slice
	models := make([]ModelInfo, 0, len(openAIResp.Data))
	for _, m := range openAIResp.Data {
		models = append(models, ModelInfo{
			ID:     m.ID,
			Name:   m.ID,
			Object: m.Object,
			Owner:  m.Owner,
		})
	}

	slog.Info("OpenAI models queried successfully", "provider", provider, "count", len(models))
	return QueryModelsResult{
		Success: true,
		Models:  models,
	}
}

// mapQueryNetworkError maps network errors for model query.
func mapQueryNetworkError(err error) string {
	errStr := err.Error()

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded") {
		return "连接超时，请检查网络或主机地址"
	}
	if strings.Contains(errStr, "connection refused") {
		return "无法连接到服务器，请检查主机地址和端口"
	}
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "lookup") {
		return "无法解析主机名，请检查 API 主机地址"
	}

	return fmt.Sprintf("网络错误: %v", err)
}
