package broker

import (
	"context"
	"testing"
	"time"

	"github.com/tedlilley/cc-bridge/internal/queue"
	"github.com/tedlilley/cc-bridge/internal/schema"
	"github.com/tedlilley/cc-bridge/internal/session"
)

// MockExecutor simulates Claude CLI execution for testing
type MockExecutor struct {
	responses map[string]string // sessionID -> response
	calls     []ExecuteCall
}

type ExecuteCall struct {
	SessionID string
	Message   string
	IsNew     bool
}

func (m *MockExecutor) Execute(ctx context.Context, sessionID string, message string, isNew bool) (*ExecuteResult, error) {
	m.calls = append(m.calls, ExecuteCall{sessionID, message, isNew})

	// Return canned response or default
	response := "mock response"
	if r, ok := m.responses[sessionID]; ok {
		response = r
	}

	newSessionID := sessionID
	if isNew {
		newSessionID = "new-session-" + sessionID
	}

	return &ExecuteResult{
		SessionID: newSessionID,
		Response:  response,
		Cost:      0.001,
	}, nil
}

func TestNewBroker(t *testing.T) {
	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	b, err := NewBroker(qMgr, sMgr, &MockExecutor{})
	if err != nil {
		t.Fatalf("NewBroker failed: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil Broker")
	}
}

func TestInitializeAgent(t *testing.T) {
	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	b, _ := NewBroker(qMgr, sMgr, &MockExecutor{})

	err := b.InitializeAgent(schema.AgentA)
	if err != nil {
		t.Fatalf("InitializeAgent failed: %v", err)
	}

	// Verify session was created
	sess, err := sMgr.GetSession(schema.AgentA)
	if err != nil {
		t.Fatalf("Session not created: %v", err)
	}
	if sess.AgentID != schema.AgentA {
		t.Errorf("expected AgentID=%q, got %q", schema.AgentA, sess.AgentID)
	}
}

func TestSendMessage(t *testing.T) {
	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	b, _ := NewBroker(qMgr, sMgr, &MockExecutor{})
	b.InitializeAgent(schema.AgentA)

	msg := schema.NewUserMessage(schema.AgentA, "hello agent")
	err := b.SendMessage(msg)
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Verify message in queue
	q, _ := qMgr.GetQueue(schema.AgentA)
	n, _ := q.Len()
	if n != 1 {
		t.Errorf("expected 1 message in queue, got %d", n)
	}
}

func TestProcessMessage(t *testing.T) {
	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	executor := &MockExecutor{
		responses: map[string]string{
			"": "I received your message",
		},
	}

	b, _ := NewBroker(qMgr, sMgr, executor)
	b.InitializeAgent(schema.AgentA)

	// Queue a message
	msg := schema.NewUserMessage(schema.AgentA, "hello")
	b.SendMessage(msg)

	// Process it
	response, err := b.ProcessNext(context.Background(), schema.AgentA)
	if err != nil {
		t.Fatalf("ProcessNext failed: %v", err)
	}
	if response == nil {
		t.Fatal("expected non-nil response")
	}
	if response.Payload.Text != "I received your message" {
		t.Errorf("unexpected response: %q", response.Payload.Text)
	}

	// Queue should be empty now
	q, _ := qMgr.GetQueue(schema.AgentA)
	n, _ := q.Len()
	if n != 0 {
		t.Errorf("expected empty queue, got %d", n)
	}
}

func TestProcessNext_EmptyQueue(t *testing.T) {
	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	b, _ := NewBroker(qMgr, sMgr, &MockExecutor{})
	b.InitializeAgent(schema.AgentA)

	// No messages queued
	response, err := b.ProcessNext(context.Background(), schema.AgentA)
	if err != nil {
		t.Fatalf("ProcessNext on empty queue should not error: %v", err)
	}
	if response != nil {
		t.Error("expected nil response for empty queue")
	}
}

func TestSessionIDUpdatedAfterFirstCall(t *testing.T) {
	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	executor := &MockExecutor{}
	b, _ := NewBroker(qMgr, sMgr, executor)
	b.InitializeAgent(schema.AgentA)

	msg := schema.NewUserMessage(schema.AgentA, "init")
	b.SendMessage(msg)
	b.ProcessNext(context.Background(), schema.AgentA)

	// Session should have ID now
	sess, _ := sMgr.GetSession(schema.AgentA)
	if sess.SessionID == "" {
		t.Error("expected SessionID to be set after first call")
	}
}

func TestTurnIncremented(t *testing.T) {
	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	executor := &MockExecutor{}
	b, _ := NewBroker(qMgr, sMgr, executor)
	b.InitializeAgent(schema.AgentA)

	// Process two messages
	for i := 0; i < 2; i++ {
		msg := schema.NewUserMessage(schema.AgentA, "msg")
		b.SendMessage(msg)
		b.ProcessNext(context.Background(), schema.AgentA)
	}

	sess, _ := sMgr.GetSession(schema.AgentA)
	if sess.TurnNumber != 2 {
		t.Errorf("expected TurnNumber=2, got %d", sess.TurnNumber)
	}
}

func TestInjectMessage(t *testing.T) {
	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	b, _ := NewBroker(qMgr, sMgr, &MockExecutor{})
	b.InitializeAgent(schema.AgentA)
	b.InitializeAgent(schema.AgentB)

	// Human injects as AgentA to AgentB
	err := b.Inject(schema.AgentA, schema.AgentB, "pretending to be agent-a")
	if err != nil {
		t.Fatalf("Inject failed: %v", err)
	}

	// Message should be in AgentB's queue
	q, _ := qMgr.GetQueue(schema.AgentB)
	msgs, _ := q.List()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].From != schema.AgentA {
		t.Errorf("expected From=%q, got %q", schema.AgentA, msgs[0].From)
	}
	if msgs[0].Type != schema.TypeInject {
		t.Errorf("expected Type=%q, got %q", schema.TypeInject, msgs[0].Type)
	}
}

func TestPollLoop(t *testing.T) {
	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	executor := &MockExecutor{
		responses: map[string]string{
			"": "polled response",
		},
	}

	b, _ := NewBroker(qMgr, sMgr, executor)
	b.InitializeAgent(schema.AgentA)

	// Queue message
	b.SendMessage(schema.NewUserMessage(schema.AgentA, "poll me"))

	// Start poll loop
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	responseCh := make(chan *schema.Message, 10)
	b.SetResponseHandler(func(msg *schema.Message) {
		responseCh <- msg
	})

	go b.Run(ctx, 50*time.Millisecond)

	// Wait for response
	select {
	case resp := <-responseCh:
		if resp.Payload.Text != "polled response" {
			t.Errorf("unexpected response: %q", resp.Payload.Text)
		}
	case <-ctx.Done():
		t.Error("timeout waiting for response")
	}
}
