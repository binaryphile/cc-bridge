package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/tedlilley/cc-bridge/internal/queue"
	"github.com/tedlilley/cc-bridge/internal/schema"
	"github.com/tedlilley/cc-bridge/internal/session"
)

// Executor interface for Claude CLI execution (allows mocking)
type Executor interface {
	Execute(ctx context.Context, sessionID string, message string, isNew bool) (*ExecuteResult, error)
}

// ExecuteResult contains the result of executing a Claude CLI command
type ExecuteResult struct {
	SessionID string
	Response  string
	Cost      float64
}

// ResponseHandler is called when a response is received
type ResponseHandler func(msg *schema.Message)

// ErrorHandler is called when an error occurs during processing
type ErrorHandler func(agent string, err error)

// Broker coordinates message passing between agents
type Broker struct {
	queueMgr     *queue.Manager
	sessionMgr   *session.Manager
	executor     Executor
	handler      ResponseHandler
	errorHandler ErrorHandler
	agents       []string
}

// NewBroker creates a new broker
func NewBroker(qMgr *queue.Manager, sMgr *session.Manager, exec Executor) (*Broker, error) {
	return &Broker{
		queueMgr:   qMgr,
		sessionMgr: sMgr,
		executor:   exec,
	}, nil
}

// InitializeAgent creates a session and queue for an agent
func (b *Broker) InitializeAgent(agentID string) error {
	_, err := b.sessionMgr.CreateSession(agentID)
	if err != nil {
		return fmt.Errorf("failed to create session for %s: %w", agentID, err)
	}

	_, err = b.queueMgr.GetQueue(agentID)
	if err != nil {
		return fmt.Errorf("failed to create queue for %s: %w", agentID, err)
	}

	// Track this agent for polling
	b.agents = append(b.agents, agentID)
	return nil
}

// SetErrorHandler sets the callback for errors
func (b *Broker) SetErrorHandler(handler ErrorHandler) {
	b.errorHandler = handler
}

// Agents returns the list of registered agents
func (b *Broker) Agents() []string {
	return b.agents
}

// SendMessage sends a message to an agent's queue
func (b *Broker) SendMessage(msg *schema.Message) error {
	q, err := b.queueMgr.GetQueue(msg.To)
	if err != nil {
		return fmt.Errorf("failed to get queue for %s: %w", msg.To, err)
	}
	return q.Enqueue(msg)
}

// ProcessNext processes the next message for an agent
func (b *Broker) ProcessNext(ctx context.Context, agentID string) (*schema.Message, error) {
	q, err := b.queueMgr.GetQueue(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue: %w", err)
	}

	msg, err := q.Dequeue()
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue: %w", err)
	}
	if msg == nil {
		return nil, nil // No messages
	}

	sess, err := b.sessionMgr.GetSession(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	isNew := sess.SessionID == ""
	result, err := b.executor.Execute(ctx, sess.SessionID, msg.Payload.Text, isNew)
	if err != nil {
		return nil, fmt.Errorf("failed to execute: %w", err)
	}

	// Update session
	if isNew || result.SessionID != sess.SessionID {
		b.sessionMgr.SetSessionID(agentID, result.SessionID)
	}
	b.sessionMgr.IncrementTurn(agentID)

	// Create response message
	response := schema.NewAgentMessage(agentID, msg.From, result.Response)
	response.WithContext(result.SessionID, sess.TurnNumber+1)
	response.WithMetadata("cost", fmt.Sprintf("%.6f", result.Cost))

	return response, nil
}

// Inject sends a message as one agent to another
func (b *Broker) Inject(asAgent, toAgent, text string) error {
	msg := schema.NewMessage(asAgent, toAgent, schema.TypeInject, text)
	return b.SendMessage(msg)
}

// SetResponseHandler sets the callback for responses
func (b *Broker) SetResponseHandler(handler ResponseHandler) {
	b.handler = handler
}

// Run starts the polling loop
func (b *Broker) Run(ctx context.Context, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, agent := range b.agents {
				resp, err := b.ProcessNext(ctx, agent)
				if err != nil {
					if b.errorHandler != nil {
						b.errorHandler(agent, err)
					}
					continue
				}
				if resp != nil && b.handler != nil {
					b.handler(resp)
				}
			}
		}
	}
}
