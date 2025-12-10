package queue

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tedlilley/cc-bridge/internal/schema"
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

func TestGetQueue(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)

	q1, err := mgr.GetQueue("agent-a")
	if err != nil {
		t.Fatalf("GetQueue failed: %v", err)
	}

	// Same agent should return same queue
	q2, _ := mgr.GetQueue("agent-a")
	if q1 != q2 {
		t.Error("expected same queue instance for same agent")
	}

	// Different agent should return different queue
	q3, _ := mgr.GetQueue("agent-b")
	if q1 == q3 {
		t.Error("expected different queue for different agent")
	}

	// Check directory was created
	if _, err := os.Stat(filepath.Join(dir, "agent-a")); os.IsNotExist(err) {
		t.Error("expected agent-a directory to exist")
	}
}

func TestEnqueueDequeue(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)
	q, _ := mgr.GetQueue("test")

	msg := schema.NewMessage("agent-a", "agent-b", schema.TypeMessage, "hello")
	if err := q.Enqueue(msg); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	got, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}

	if got.ID != msg.ID {
		t.Errorf("ID mismatch: %q != %q", got.ID, msg.ID)
	}
	if got.Payload.Text != "hello" {
		t.Errorf("Payload.Text mismatch: %q", got.Payload.Text)
	}
}

func TestDequeueEmpty(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)
	q, _ := mgr.GetQueue("test")

	msg, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue on empty queue should not error: %v", err)
	}
	if msg != nil {
		t.Error("expected nil message from empty queue")
	}
}

func TestFIFOOrder(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)
	q, _ := mgr.GetQueue("test")

	// Enqueue in order
	for i := 1; i <= 3; i++ {
		msg := schema.NewMessage("a", "b", schema.TypeMessage, string(rune('0'+i)))
		time.Sleep(time.Millisecond) // Ensure different timestamps
		q.Enqueue(msg)
	}

	// Dequeue should be FIFO
	for i := 1; i <= 3; i++ {
		msg, _ := q.Dequeue()
		expected := string(rune('0' + i))
		if msg.Payload.Text != expected {
			t.Errorf("expected %q, got %q", expected, msg.Payload.Text)
		}
	}
}

func TestPeek(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)
	q, _ := mgr.GetQueue("test")

	msg := schema.NewMessage("a", "b", schema.TypeMessage, "peek-test")
	q.Enqueue(msg)

	// Peek should return message without removing
	peeked, err := q.Peek()
	if err != nil {
		t.Fatalf("Peek failed: %v", err)
	}
	if peeked.ID != msg.ID {
		t.Error("Peek returned wrong message")
	}

	// Peek again should return same message
	peeked2, _ := q.Peek()
	if peeked2.ID != msg.ID {
		t.Error("second Peek returned different message")
	}

	// Dequeue should still work
	dequeued, _ := q.Dequeue()
	if dequeued.ID != msg.ID {
		t.Error("Dequeue after Peek returned wrong message")
	}
}

func TestPeekEmpty(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)
	q, _ := mgr.GetQueue("test")

	msg, err := q.Peek()
	if err != nil {
		t.Fatalf("Peek on empty queue should not error: %v", err)
	}
	if msg != nil {
		t.Error("expected nil from Peek on empty queue")
	}
}

func TestLen(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)
	q, _ := mgr.GetQueue("test")

	n, _ := q.Len()
	if n != 0 {
		t.Errorf("expected Len=0, got %d", n)
	}

	q.Enqueue(schema.NewMessage("a", "b", schema.TypeMessage, "1"))
	q.Enqueue(schema.NewMessage("a", "b", schema.TypeMessage, "2"))

	n, _ = q.Len()
	if n != 2 {
		t.Errorf("expected Len=2, got %d", n)
	}

	q.Dequeue()
	n, _ = q.Len()
	if n != 1 {
		t.Errorf("expected Len=1, got %d", n)
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)
	q, _ := mgr.GetQueue("test")

	q.Enqueue(schema.NewMessage("a", "b", schema.TypeMessage, "first"))
	time.Sleep(time.Millisecond)
	q.Enqueue(schema.NewMessage("a", "b", schema.TypeMessage, "second"))

	msgs, err := q.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Payload.Text != "first" {
		t.Errorf("expected first message, got %q", msgs[0].Payload.Text)
	}
	if msgs[1].Payload.Text != "second" {
		t.Errorf("expected second message, got %q", msgs[1].Payload.Text)
	}

	// List should not remove messages
	n, _ := q.Len()
	if n != 2 {
		t.Error("List should not remove messages")
	}
}

func TestClear(t *testing.T) {
	dir := t.TempDir()
	mgr, _ := NewManager(dir)
	q, _ := mgr.GetQueue("test")

	q.Enqueue(schema.NewMessage("a", "b", schema.TypeMessage, "1"))
	q.Enqueue(schema.NewMessage("a", "b", schema.TypeMessage, "2"))

	if err := q.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	n, _ := q.Len()
	if n != 0 {
		t.Errorf("expected Len=0 after Clear, got %d", n)
	}
}
