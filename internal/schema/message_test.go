package schema

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage("agent-a", "agent-b", TypeMessage, "hello")

	if msg.ID == "" {
		t.Error("expected non-empty ID")
	}
	if msg.From != "agent-a" {
		t.Errorf("expected From='agent-a', got %q", msg.From)
	}
	if msg.To != "agent-b" {
		t.Errorf("expected To='agent-b', got %q", msg.To)
	}
	if msg.Type != TypeMessage {
		t.Errorf("expected Type=%q, got %q", TypeMessage, msg.Type)
	}
	if msg.Payload.Text != "hello" {
		t.Errorf("expected Payload.Text='hello', got %q", msg.Payload.Text)
	}
	if msg.Timestamp.IsZero() {
		t.Error("expected non-zero Timestamp")
	}
}

func TestNewUserMessage(t *testing.T) {
	msg := NewUserMessage(AgentA, "test message")

	if msg.From != Human {
		t.Errorf("expected From=%q, got %q", Human, msg.From)
	}
	if msg.To != AgentA {
		t.Errorf("expected To=%q, got %q", AgentA, msg.To)
	}
}

func TestNewAgentMessage(t *testing.T) {
	msg := NewAgentMessage(AgentA, AgentB, "agent reply")

	if msg.From != AgentA {
		t.Errorf("expected From=%q, got %q", AgentA, msg.From)
	}
	if msg.To != AgentB {
		t.Errorf("expected To=%q, got %q", AgentB, msg.To)
	}
}

func TestWithMetadata(t *testing.T) {
	msg := NewMessage(AgentA, AgentB, TypeMessage, "test").
		WithMetadata("key1", "value1").
		WithMetadata("key2", "value2")

	if msg.Payload.Metadata["key1"] != "value1" {
		t.Errorf("expected Metadata[key1]='value1', got %q", msg.Payload.Metadata["key1"])
	}
	if msg.Payload.Metadata["key2"] != "value2" {
		t.Errorf("expected Metadata[key2]='value2', got %q", msg.Payload.Metadata["key2"])
	}
}

func TestWithContext(t *testing.T) {
	msg := NewMessage(AgentA, AgentB, TypeMessage, "test").
		WithContext("session-123", 5)

	if msg.Context == nil {
		t.Fatal("expected non-nil Context")
	}
	if msg.Context.SessionID != "session-123" {
		t.Errorf("expected SessionID='session-123', got %q", msg.Context.SessionID)
	}
	if msg.Context.TurnNumber != 5 {
		t.Errorf("expected TurnNumber=5, got %d", msg.Context.TurnNumber)
	}
}

func TestToJSON(t *testing.T) {
	msg := NewMessage(AgentA, AgentB, TypeMessage, "test")
	data, err := msg.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if decoded["from"] != AgentA {
		t.Errorf("expected from=%q, got %v", AgentA, decoded["from"])
	}
}

func TestFromJSON(t *testing.T) {
	original := NewMessage(AgentA, AgentB, TypeMessage, "test").
		WithMetadata("foo", "bar").
		WithContext("sess-1", 3)

	data, _ := original.ToJSON()
	restored, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("ID mismatch: %q != %q", restored.ID, original.ID)
	}
	if restored.From != original.From {
		t.Errorf("From mismatch: %q != %q", restored.From, original.From)
	}
	if restored.Payload.Text != original.Payload.Text {
		t.Errorf("Payload.Text mismatch: %q != %q", restored.Payload.Text, original.Payload.Text)
	}
	if restored.Payload.Metadata["foo"] != "bar" {
		t.Errorf("Metadata mismatch: got %v", restored.Payload.Metadata)
	}
	if restored.Context.SessionID != "sess-1" {
		t.Errorf("Context.SessionID mismatch: %q", restored.Context.SessionID)
	}
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	_, err := FromJSON([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are defined
	if AgentA != "agent-a" {
		t.Errorf("AgentA=%q, expected 'agent-a'", AgentA)
	}
	if AgentB != "agent-b" {
		t.Errorf("AgentB=%q, expected 'agent-b'", AgentB)
	}
	if Human != "human" {
		t.Errorf("Human=%q, expected 'human'", Human)
	}
	if Broadcast != "broadcast" {
		t.Errorf("Broadcast=%q, expected 'broadcast'", Broadcast)
	}
	if TypeMessage != "message" {
		t.Errorf("TypeMessage=%q, expected 'message'", TypeMessage)
	}
	if TypeToolResult != "tool_result" {
		t.Errorf("TypeToolResult=%q, expected 'tool_result'", TypeToolResult)
	}
	if TypeInject != "inject" {
		t.Errorf("TypeInject=%q, expected 'inject'", TypeInject)
	}
	if TypeSystem != "system" {
		t.Errorf("TypeSystem=%q, expected 'system'", TypeSystem)
	}
}

func TestTimestampIsUTC(t *testing.T) {
	before := time.Now().UTC()
	msg := NewMessage(AgentA, AgentB, TypeMessage, "test")
	after := time.Now().UTC()

	if msg.Timestamp.Before(before) || msg.Timestamp.After(after) {
		t.Errorf("Timestamp %v not between %v and %v", msg.Timestamp, before, after)
	}
	if msg.Timestamp.Location() != time.UTC {
		t.Errorf("expected UTC timezone, got %v", msg.Timestamp.Location())
	}
}
