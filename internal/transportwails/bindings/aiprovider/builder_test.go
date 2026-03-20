package aiprovider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeURLPreservesHostBasePath(t *testing.T) {
	got, err := NormalizeURL("https://open.bigmodel.cn/api/paas/v4", "/chat/completions")
	if err != nil {
		t.Fatalf("NormalizeURL returned error: %v", err)
	}
	if got != "https://open.bigmodel.cn/api/paas/v4/chat/completions" {
		t.Fatalf("unexpected normalized URL: %s", got)
	}
}

func TestGLMTestConnectionUsesProviderBasePath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/paas/v4/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["model"] != "glm-4-flash" {
			t.Fatalf("unexpected model: %#v", payload["model"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer server.Close()

	result := TestConnection(context.Background(), TestRequest{
		Provider:    "glm",
		APIHost:     server.URL + "/api/paas/v4",
		APIEndpoint: "/chat/completions",
		APIKey:      "glm-key",
		Model:       "glm-4-flash",
	})

	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
}

func TestMiniMaxSendChatUsesOfficialCompatibleEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer mm-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["model"] != "MiniMax-M2.7" {
			t.Fatalf("unexpected model: %#v", payload["model"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello from minimax"}}]}`))
	}))
	defer server.Close()

	result := SendChat(context.Background(), ChatRequest{
		Provider:    "minimax",
		APIHost:     server.URL,
		APIEndpoint: "/v1/chat/completions",
		APIKey:      "mm-key",
		Model:       "MiniMax-M2.7",
		Prompt:      "hello",
		Temperature: 0.1,
	})

	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Content != "hello from minimax" {
		t.Fatalf("unexpected content: %q", result.Content)
	}
}

func TestMiniMaxChinaDefaultURLAndOpenAIBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer mm-cn-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("unexpected content-type: %s", got)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["model"] != "MiniMax-M2.7" {
			t.Fatalf("unexpected model: %#v", payload["model"])
		}
		messages, ok := payload["messages"].([]any)
		if !ok || len(messages) != 1 {
			t.Fatalf("unexpected messages payload: %#v", payload["messages"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"china minimax ok"}}]}`))
	}))
	defer server.Close()

	result := SendChat(context.Background(), ChatRequest{
		Provider: "minimax",
		APIHost:  server.URL,
		APIKey:   "mm-cn-key",
		Model:    "MiniMax-M2.7",
		Prompt:   "hello",
	})

	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Content != "china minimax ok" {
		t.Fatalf("unexpected content: %q", result.Content)
	}
}

func TestMiniMaxTestConnectionUsesOfficialCompatibleEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer server.Close()

	result := TestConnection(context.Background(), TestRequest{
		Provider:    "minimax",
		APIHost:     server.URL,
		APIEndpoint: "/v1/chat/completions",
		APIKey:      "mm-key",
		Model:       "MiniMax-M2.7",
	})

	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
}

func TestQueryModelsUnsupportedProviderStillAllowsManualModelChat(t *testing.T) {
	queryResult := QueryModels(context.Background(), QueryModelsRequest{
		Provider: "glm",
		APIHost:  "https://open.bigmodel.cn/api/paas/v4",
		APIKey:   "glm-key",
	})
	if queryResult.Success {
		t.Fatal("expected GLM model query to remain unsupported")
	}
	if !strings.Contains(queryResult.Error, "手动输入模型名称") {
		t.Fatalf("unexpected query error: %s", queryResult.Error)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/paas/v4/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"manual glm ok"}}]}`))
	}))
	defer server.Close()

	chatResult := SendChat(context.Background(), ChatRequest{
		Provider:    "glm",
		APIHost:     server.URL + "/api/paas/v4",
		APIEndpoint: "/chat/completions",
		APIKey:      "glm-key",
		Model:       "glm-4-flash",
		Prompt:      "hello",
		Temperature: 0.1,
	})
	if !chatResult.Success {
		t.Fatalf("expected manual-model chat to succeed, got error: %s", chatResult.Error)
	}
}

func TestSendChatParsesOpenAIResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Fatalf("unexpected auth header: %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello from deepseek"}}]}`))
	}))
	defer server.Close()

	result := SendChat(context.Background(), ChatRequest{
		Provider:    "deepseek",
		APIHost:     server.URL,
		APIEndpoint: "/v1/chat/completions",
		APIKey:      "sk-test",
		Model:       "deepseek-chat",
		Prompt:      "你是谁？请简单介绍你自己。",
		Temperature: 0.1,
	})

	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Content != "hello from deepseek" {
		t.Fatalf("unexpected content: %q", result.Content)
	}
}

func TestSendChatParsesOllamaResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"content":"hello from ollama"}}`))
	}))
	defer server.Close()

	result := SendChat(context.Background(), ChatRequest{
		Provider:    "ollama",
		APIHost:     server.URL,
		APIEndpoint: "/api/chat",
		Model:       "llama3",
		Prompt:      "你是谁？请简单介绍你自己。",
		Temperature: 0.1,
	})

	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Content != "hello from ollama" {
		t.Fatalf("unexpected content: %q", result.Content)
	}
}

func TestSendChatRejectsEmptyPrompt(t *testing.T) {
	result := SendChat(context.Background(), ChatRequest{
		Provider: "deepseek",
		APIHost:  "https://api.deepseek.com",
		APIKey:   "sk-test",
		Model:    "deepseek-chat",
		Prompt:   "   ",
	})

	if result.Success {
		t.Fatal("expected empty prompt to fail")
	}
	if !strings.Contains(result.Error, "测试问题不能为空") {
		t.Fatalf("unexpected error: %s", result.Error)
	}
}

func TestTestConnectionFailureIncludesAuditContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"error":"method not allowed for this endpoint"}`))
	}))
	defer server.Close()

	result := TestConnection(context.Background(), TestRequest{
		Provider:    "minimax",
		APIHost:     server.URL,
		APIEndpoint: "/v1/chat/completions",
		APIKey:      "mm-secret-key",
		Model:       "MiniMax-M2.7",
	})

	if result.Success {
		t.Fatal("expected failure")
	}
	for _, want := range []string{
		"测试连接失败",
		"POST " + server.URL + "/v1/chat/completions",
		"HTTP 405",
		`{"error":"method not allowed for this endpoint"}`,
	} {
		if !strings.Contains(result.Error, want) {
			t.Fatalf("expected error to contain %q, got %q", want, result.Error)
		}
	}
}

func TestSendChatFailureIncludesAuditContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"temperature is not supported"}`))
	}))
	defer server.Close()

	result := SendChat(context.Background(), ChatRequest{
		Provider:    "glm",
		APIHost:     server.URL + "/api/paas/v4",
		APIEndpoint: "/chat/completions",
		APIKey:      "glm-secret-key",
		Model:       "glm-4-flash",
		Prompt:      "hello",
		Temperature: 0.1,
	})

	if result.Success {
		t.Fatal("expected failure")
	}
	for _, want := range []string{
		"测试模型失败",
		"POST " + server.URL + "/api/paas/v4/chat/completions",
		"HTTP 400",
		`{"error":"temperature is not supported"}`,
	} {
		if !strings.Contains(result.Error, want) {
			t.Fatalf("expected error to contain %q, got %q", want, result.Error)
		}
	}
}

func TestMaskHeadersRedactsAuthorization(t *testing.T) {
	headers := maskHeaders(http.Header{
		"Authorization": []string{"Bearer super-secret-token"},
		"Content-Type":  []string{"application/json"},
	})

	if got := headers["Authorization"]; got == "Bearer super-secret-token" {
		t.Fatalf("expected authorization to be redacted, got %q", got)
	}
	if got := headers["Content-Type"]; got != "application/json" {
		t.Fatalf("unexpected content-type header: %q", got)
	}
}

func TestMapAPIErrorExtractsNestedProviderMessage(t *testing.T) {
	body := []byte(`{"type":"error","error":{"type":"authorized_error","message":"login fail: Please carry the API secret key in the Authorization header","http_code":"401"}}`)

	got := mapAPIError(http.StatusBadRequest, body, "minimax")

	if !strings.Contains(got, "login fail") {
		t.Fatalf("expected nested provider message, got %q", got)
	}
}

func TestMapAPIErrorIncludesProviderMessageFor401(t *testing.T) {
	body := []byte(`{"type":"error","error":{"type":"authorized_error","message":"login fail: invalid api key","http_code":"401"}}`)

	got := mapAPIError(http.StatusUnauthorized, body, "minimax")

	if !strings.Contains(got, "invalid api key") {
		t.Fatalf("expected 401 error to include provider message, got %q", got)
	}
}

func TestMapAPIErrorIncludesMiniMax2049ProviderMessage(t *testing.T) {
	body := []byte(`{"type":"error","error":{"type":"authorized_error","message":"invalid api key (2049)","http_code":"401"}}`)

	got := mapAPIError(http.StatusUnauthorized, body, "minimax")

	if !strings.Contains(got, "invalid api key (2049)") {
		t.Fatalf("expected MiniMax provider error details, got %q", got)
	}
}
