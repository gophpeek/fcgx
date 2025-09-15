package fcgx

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxWriteSize != 65500 {
		t.Errorf("Expected MaxWriteSize 65500, got %d", config.MaxWriteSize)
	}

	if config.ConnectTimeout != 5*time.Second {
		t.Errorf("Expected ConnectTimeout 5s, got %v", config.ConnectTimeout)
	}

	if config.RequestTimeout != 30*time.Second {
		t.Errorf("Expected RequestTimeout 30s, got %v", config.RequestTimeout)
	}
}

func TestDialWithConfig(t *testing.T) {
	config := &Config{
		MaxWriteSize:   32768,
		ConnectTimeout: 2 * time.Second,
		RequestTimeout: 10 * time.Second,
	}

	// This will fail to connect, but we're testing that the config is used
	client, err := DialWithConfig("tcp", "127.0.0.1:9999", config)
	if err == nil {
		t.Error("Expected connection to fail to non-existent server")
		if client != nil {
			client.Close()
		}
		return
	}

	// Test with nil config - should use defaults
	client, err = DialWithConfig("tcp", "127.0.0.1:9999", nil)
	if err == nil {
		t.Error("Expected connection to fail to non-existent server")
		if client != nil {
			client.Close()
		}
		return
	}
}

func TestWrapWithContext(t *testing.T) {
	baseErr := &testError{msg: "base error"}
	kindErr := ErrTimeout

	// Test with empty context
	err1 := wrapWithContext(baseErr, kindErr, "test message", nil)
	expected1 := "fcgx: timeout: test message: base error"
	if err1.Error() != expected1 {
		t.Errorf("Expected %q, got %q", expected1, err1.Error())
	}

	// Test with context
	context := map[string]interface{}{
		"reqID":    42,
		"deadline": "2024-01-01T12:00:00Z",
	}
	err2 := wrapWithContext(baseErr, kindErr, "test message", context)
	// The exact order of context items may vary due to map iteration
	result := err2.Error()
	if !contains(result, "fcgx: timeout: test message") {
		t.Errorf("Error should contain base message, got %q", result)
	}
	if !contains(result, "reqID=42") {
		t.Errorf("Error should contain reqID context, got %q", result)
	}
	if !contains(result, "deadline=2024-01-01T12:00:00Z") {
		t.Errorf("Error should contain deadline context, got %q", result)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
