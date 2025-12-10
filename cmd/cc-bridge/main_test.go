package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgs_Start(t *testing.T) {
	args := []string{"start"}
	cmd, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	if cmd.Command != "start" {
		t.Errorf("expected command='start', got %q", cmd.Command)
	}
}

func TestParseArgs_StartWithDir(t *testing.T) {
	args := []string{"start", "--data-dir", "/tmp/test"}
	cmd, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	if cmd.DataDir != "/tmp/test" {
		t.Errorf("expected DataDir='/tmp/test', got %q", cmd.DataDir)
	}
}

func TestParseArgs_StartWithInterval(t *testing.T) {
	args := []string{"start", "--poll-interval", "500ms"}
	cmd, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	if cmd.PollInterval.Milliseconds() != 500 {
		t.Errorf("expected PollInterval=500ms, got %v", cmd.PollInterval)
	}
}

func TestParseArgs_Status(t *testing.T) {
	args := []string{"status"}
	cmd, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	if cmd.Command != "status" {
		t.Errorf("expected command='status', got %q", cmd.Command)
	}
}

func TestParseArgs_Send(t *testing.T) {
	args := []string{"send", "--to", "agent-a", "hello there"}
	cmd, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	if cmd.Command != "send" {
		t.Errorf("expected command='send', got %q", cmd.Command)
	}
	if cmd.To != "agent-a" {
		t.Errorf("expected To='agent-a', got %q", cmd.To)
	}
	if cmd.Message != "hello there" {
		t.Errorf("expected Message='hello there', got %q", cmd.Message)
	}
}

func TestParseArgs_Inject(t *testing.T) {
	args := []string{"inject", "--as", "agent-a", "--to", "agent-b", "injected message"}
	cmd, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	if cmd.Command != "inject" {
		t.Errorf("expected command='inject', got %q", cmd.Command)
	}
	if cmd.As != "agent-a" {
		t.Errorf("expected As='agent-a', got %q", cmd.As)
	}
	if cmd.To != "agent-b" {
		t.Errorf("expected To='agent-b', got %q", cmd.To)
	}
	if cmd.Message != "injected message" {
		t.Errorf("expected Message='injected message', got %q", cmd.Message)
	}
}

func TestParseArgs_NoCommand(t *testing.T) {
	args := []string{}
	_, err := ParseArgs(args)
	if err == nil {
		t.Error("expected error for no command")
	}
}

func TestParseArgs_InvalidCommand(t *testing.T) {
	args := []string{"invalid"}
	_, err := ParseArgs(args)
	if err == nil {
		t.Error("expected error for invalid command")
	}
}

func TestDefaultDataDir(t *testing.T) {
	dir := DefaultDataDir()
	if dir == "" {
		t.Error("expected non-empty default data dir")
	}

	// Should be under home or /tmp
	home, _ := os.UserHomeDir()
	if !filepath.HasPrefix(dir, home) && !filepath.HasPrefix(dir, "/tmp") {
		t.Errorf("unexpected data dir: %s", dir)
	}
}
