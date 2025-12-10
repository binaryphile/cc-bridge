package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// ClaudeExecutor executes Claude CLI commands
type ClaudeExecutor struct{}

// NewClaudeExecutor creates a new ClaudeExecutor
func NewClaudeExecutor() *ClaudeExecutor {
	return &ClaudeExecutor{}
}

// claudeOutput represents the JSON output from claude CLI
type claudeOutput struct {
	SessionID string  `json:"session_id"`
	Result    string  `json:"result"`
	Cost      float64 `json:"total_cost_usd"`
}

// BuildArgs builds the command line arguments for claude
func (e *ClaudeExecutor) BuildArgs(sessionID, message string, isNew bool) []string {
	args := []string{}

	if isNew {
		args = append(args, "-p", message)
	} else {
		args = append(args, "--resume", sessionID, "-p", message)
	}

	args = append(args, "--output-format", "json", "--max-turns", "1")
	return args
}

// ParseResult parses the JSON output from claude
func (e *ClaudeExecutor) ParseResult(output []byte) (*ExecuteResult, error) {
	var co claudeOutput
	if err := json.Unmarshal(output, &co); err != nil {
		return nil, fmt.Errorf("failed to parse claude output: %w", err)
	}

	return &ExecuteResult{
		SessionID: co.SessionID,
		Response:  co.Result,
		Cost:      co.Cost,
	}, nil
}

// Execute runs the claude CLI and returns the result
func (e *ClaudeExecutor) Execute(ctx context.Context, sessionID, message string, isNew bool) (*ExecuteResult, error) {
	args := e.BuildArgs(sessionID, message, isNew)

	cmd := exec.CommandContext(ctx, "claude", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("claude command failed: %w, stderr: %s", err, stderr.String())
	}

	return e.ParseResult(stdout.Bytes())
}
