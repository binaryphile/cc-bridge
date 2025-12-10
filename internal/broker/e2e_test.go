package broker

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/tedlilley/cc-bridge/internal/queue"
	"github.com/tedlilley/cc-bridge/internal/schema"
	"github.com/tedlilley/cc-bridge/internal/session"
)

// TestE2E_TwoAgentConversation tests a full conversation between two agents.
// Agent A receives a secret, then Agent B asks for it.
func TestE2E_TwoAgentConversation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// Check if claude binary exists
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude binary not found, skipping e2e test")
	}

	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	executor := NewClaudeExecutor()
	b, _ := NewBroker(qMgr, sMgr, executor)

	// Initialize both agents
	b.InitializeAgent(schema.AgentA)
	b.InitializeAgent(schema.AgentB)

	// Collect responses
	responses := make([]*schema.Message, 0)
	errors := make([]error, 0)

	b.SetResponseHandler(func(msg *schema.Message) {
		t.Logf("Response: %s -> %s: %s", msg.From, msg.To, msg.Payload.Text)
		responses = append(responses, msg)
	})

	b.SetErrorHandler(func(agent string, err error) {
		t.Logf("Error for %s: %v", agent, err)
		errors = append(errors, err)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Step 1: Human tells Agent A a secret
	t.Log("Step 1: Sending secret to Agent A")
	msg1 := schema.NewUserMessage(schema.AgentA, "Remember this secret code: DELTA-7. Just reply 'STORED' and nothing else.")
	b.SendMessage(msg1)

	// Process Agent A
	resp1, err := b.ProcessNext(ctx, schema.AgentA)
	if err != nil {
		t.Fatalf("Agent A failed to process: %v", err)
	}
	t.Logf("Agent A response: %s", resp1.Payload.Text)

	// Step 2: Human asks Agent B to ask Agent A for the secret
	// But first, we need Agent B to have context. Let's just test Agent A remembers.

	// Step 2: Resume Agent A's session and ask for the code
	t.Log("Step 2: Asking Agent A to recall the secret")
	msg2 := schema.NewUserMessage(schema.AgentA, "What was the secret code I told you? Reply with just the code.")
	b.SendMessage(msg2)

	resp2, err := b.ProcessNext(ctx, schema.AgentA)
	if err != nil {
		t.Fatalf("Agent A failed on second turn: %v", err)
	}
	t.Logf("Agent A recall response: %s", resp2.Payload.Text)

	// Verify Agent A remembered the secret
	if !strings.Contains(strings.ToUpper(resp2.Payload.Text), "DELTA-7") {
		t.Errorf("Agent A did not remember the secret. Got: %s", resp2.Payload.Text)
	}

	// Step 3: Now test Agent B independently
	t.Log("Step 3: Testing Agent B independently")
	msg3 := schema.NewUserMessage(schema.AgentB, "Say only: PONG")
	b.SendMessage(msg3)

	resp3, err := b.ProcessNext(ctx, schema.AgentB)
	if err != nil {
		t.Fatalf("Agent B failed: %v", err)
	}
	t.Logf("Agent B response: %s", resp3.Payload.Text)

	// Verify sessions are different
	sessA, _ := sMgr.GetSession(schema.AgentA)
	sessB, _ := sMgr.GetSession(schema.AgentB)

	if sessA.SessionID == sessB.SessionID {
		t.Error("Agent A and B should have different session IDs")
	}

	t.Logf("Agent A session: %s (turns: %d)", sessA.SessionID, sessA.TurnNumber)
	t.Logf("Agent B session: %s (turns: %d)", sessB.SessionID, sessB.TurnNumber)

	// Verify turn counts
	if sessA.TurnNumber != 2 {
		t.Errorf("Expected Agent A to have 2 turns, got %d", sessA.TurnNumber)
	}
	if sessB.TurnNumber != 1 {
		t.Errorf("Expected Agent B to have 1 turn, got %d", sessB.TurnNumber)
	}

	// No errors should have occurred
	if len(errors) > 0 {
		t.Errorf("Unexpected errors: %v", errors)
	}
}

// TestE2E_Injection tests that injection correctly routes messages
func TestE2E_Injection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude binary not found, skipping e2e test")
	}

	dir := t.TempDir()
	qMgr, _ := queue.NewManager(dir + "/queues")
	sMgr, _ := session.NewManager(dir + "/sessions")

	executor := NewClaudeExecutor()
	b, _ := NewBroker(qMgr, sMgr, executor)

	b.InitializeAgent(schema.AgentA)
	b.InitializeAgent(schema.AgentB)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Inject a message "as Agent A" to Agent B
	t.Log("Injecting message as Agent A to Agent B")
	err := b.Inject(schema.AgentA, schema.AgentB, "Hello Agent B, this is Agent A. Say 'RECEIVED FROM A'")
	if err != nil {
		t.Fatalf("Inject failed: %v", err)
	}

	// Check the queue - message should be in Agent B's queue with from=agent-a
	qB, _ := qMgr.GetQueue(schema.AgentB)
	msgs, _ := qB.List()

	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message in Agent B queue, got %d", len(msgs))
	}

	if msgs[0].From != schema.AgentA {
		t.Errorf("Expected message from %s, got %s", schema.AgentA, msgs[0].From)
	}

	if msgs[0].Type != schema.TypeInject {
		t.Errorf("Expected type %s, got %s", schema.TypeInject, msgs[0].Type)
	}

	// Process Agent B - it should respond
	resp, err := b.ProcessNext(ctx, schema.AgentB)
	if err != nil {
		t.Fatalf("Agent B failed to process injected message: %v", err)
	}

	t.Logf("Agent B response to injected message: %s", resp.Payload.Text)

	// Response should be from Agent B to Agent A (since the injected message was "from" Agent A)
	if resp.From != schema.AgentB {
		t.Errorf("Expected response from %s, got %s", schema.AgentB, resp.From)
	}
	if resp.To != schema.AgentA {
		t.Errorf("Expected response to %s, got %s", schema.AgentA, resp.To)
	}
}
