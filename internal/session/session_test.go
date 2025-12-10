package session

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	if mgr == nil {
		t.Fatal("expected non-nil Manager")
	}
}

func TestCreateSession(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)

	sess, err := mgr.CreateSession("agent-a")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if sess.AgentID != "agent-a" {
		t.Errorf("expected AgentID='agent-a', got %q", sess.AgentID)
	}
	if sess.SessionID != "" {
		t.Error("new session should have empty SessionID until initialized")
	}
	if sess.TurnNumber != 0 {
		t.Errorf("expected TurnNumber=0, got %d", sess.TurnNumber)
	}
}

func TestGetSession(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)

	// Create session
	created, _ := mgr.CreateSession("agent-a")

	// Get it back
	got, err := mgr.GetSession("agent-a")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if got.AgentID != created.AgentID {
		t.Error("GetSession returned different session")
	}
}

func TestGetSession_NotFound(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)

	_, err := mgr.GetSession("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestSetSessionID(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)

	sess, _ := mgr.CreateSession("agent-a")
	err := mgr.SetSessionID("agent-a", "claude-session-12345")
	if err != nil {
		t.Fatalf("SetSessionID failed: %v", err)
	}

	// Verify it was set
	got, _ := mgr.GetSession("agent-a")
	if got.SessionID != "claude-session-12345" {
		t.Errorf("expected SessionID='claude-session-12345', got %q", got.SessionID)
	}

	// Original should also reflect change (same pointer)
	if sess.SessionID != "claude-session-12345" {
		t.Error("original session should reflect SessionID change")
	}
}

func TestIncrementTurn(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)

	mgr.CreateSession("agent-a")

	err := mgr.IncrementTurn("agent-a")
	if err != nil {
		t.Fatalf("IncrementTurn failed: %v", err)
	}

	sess, _ := mgr.GetSession("agent-a")
	if sess.TurnNumber != 1 {
		t.Errorf("expected TurnNumber=1, got %d", sess.TurnNumber)
	}

	mgr.IncrementTurn("agent-a")
	sess, _ = mgr.GetSession("agent-a")
	if sess.TurnNumber != 2 {
		t.Errorf("expected TurnNumber=2, got %d", sess.TurnNumber)
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	mgr1, _ := NewManager(dir)

	// Create and modify session
	mgr1.CreateSession("agent-a")
	mgr1.SetSessionID("agent-a", "persisted-session")
	mgr1.IncrementTurn("agent-a")
	mgr1.IncrementTurn("agent-a")

	// Save state
	if err := mgr1.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new manager from same directory
	mgr2, _ := NewManager(dir)
	if err := mgr2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify state was restored
	sess, err := mgr2.GetSession("agent-a")
	if err != nil {
		t.Fatalf("GetSession after Load failed: %v", err)
	}
	if sess.SessionID != "persisted-session" {
		t.Errorf("expected SessionID='persisted-session', got %q", sess.SessionID)
	}
	if sess.TurnNumber != 2 {
		t.Errorf("expected TurnNumber=2, got %d", sess.TurnNumber)
	}
}

func TestLoadEmpty(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)

	// Load from empty directory should not error
	err := mgr.Load()
	if err != nil {
		t.Fatalf("Load from empty dir should not error: %v", err)
	}
}

func TestListSessions(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)

	mgr.CreateSession("agent-a")
	mgr.CreateSession("agent-b")

	sessions := mgr.ListSessions()
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	// Check both agents are present
	found := make(map[string]bool)
	for _, s := range sessions {
		found[s.AgentID] = true
	}
	if !found["agent-a"] || !found["agent-b"] {
		t.Error("expected both agent-a and agent-b in list")
	}
}
