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
