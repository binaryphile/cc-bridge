package broker

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestClaudeExecutorImplementsInterface(t *testing.T) {
	// Compile-time check that ClaudeExecutor implements Executor
	var _ Executor = (*ClaudeExecutor)(nil)
}

func TestNewClaudeExecutor(t *testing.T) {
	exec := NewClaudeExecutor()
	if exec == nil {
		t.Fatal("expected non-nil ClaudeExecutor")
	}
}

func TestClaudeExecutor_BuildCommand_New(t *testing.T) {
	exec := NewClaudeExecutor()

	args := exec.BuildArgs("", "hello world", true)

	// Should have -p flag
	hasP := false
	for _, arg := range args {
		if arg == "-p" {
			hasP = true
		}
	}
	if !hasP {
		t.Error("expected -p flag for new session")
	}

	// Should NOT have --resume
	hasResume := false
	for _, arg := range args {
		if arg == "--resume" {
			hasResume = true
		}
	}
	if hasResume {
		t.Error("should not have --resume for new session")
	}

	// Should have --output-format json
	hasFormat := false
	for i, arg := range args {
		if arg == "--output-format" && i+1 < len(args) && args[i+1] == "json" {
			hasFormat = true
		}
	}
	if !hasFormat {
		t.Error("expected --output-format json")
	}

	// Should have --max-turns 1
	hasMaxTurns := false
	for i, arg := range args {
		if arg == "--max-turns" && i+1 < len(args) && args[i+1] == "1" {
			hasMaxTurns = true
		}
	}
	if !hasMaxTurns {
		t.Error("expected --max-turns 1")
	}
}

func TestClaudeExecutor_BuildCommand_Resume(t *testing.T) {
	exec := NewClaudeExecutor()

	args := exec.BuildArgs("session-123", "hello world", false)

	// Should have --resume session-123
	hasResume := false
	for i, arg := range args {
		if arg == "--resume" && i+1 < len(args) && args[i+1] == "session-123" {
			hasResume = true
		}
	}
	if !hasResume {
		t.Error("expected --resume session-123 for resume")
	}
}

func TestClaudeExecutor_ParseResult(t *testing.T) {
	exec := NewClaudeExecutor()

	jsonOutput := `{"session_id":"abc-123","result":"Hello there!","total_cost_usd":0.001234}`

	result, err := exec.ParseResult([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("ParseResult failed: %v", err)
	}

	if result.SessionID != "abc-123" {
		t.Errorf("expected SessionID='abc-123', got %q", result.SessionID)
	}
	if result.Response != "Hello there!" {
		t.Errorf("expected Response='Hello there!', got %q", result.Response)
	}
	if result.Cost != 0.001234 {
		t.Errorf("expected Cost=0.001234, got %f", result.Cost)
	}
}

func TestClaudeExecutor_ParseResult_InvalidJSON(t *testing.T) {
	exec := NewClaudeExecutor()

	_, err := exec.ParseResult([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// Integration test - only run if claude binary is available
func TestClaudeExecutor_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check if claude binary exists
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude binary not found, skipping integration test")
	}

	executor := NewClaudeExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, "", "Say only: PONG", true)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.SessionID == "" {
		t.Error("expected non-empty SessionID")
	}
	if result.Response == "" {
		t.Error("expected non-empty Response")
	}
}
