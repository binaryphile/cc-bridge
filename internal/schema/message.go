package schema

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	AgentA    = "agent-a"
	AgentB    = "agent-b"
	Human     = "human"
	Broadcast = "broadcast"
)

const (
	TypeMessage    = "message"
	TypeToolResult = "tool_result"
	TypeInject     = "inject"
	TypeSystem     = "system"
)

type Message struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Type      string    `json:"type"`
	Payload   Payload   `json:"payload"`
	Context   *Context  `json:"context,omitempty"`
}

type Payload struct {
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Context struct {
	SessionID  string `json:"session_id,omitempty"`
	TurnNumber int    `json:"turn_number,omitempty"`
}

func NewMessage(from, to, msgType, text string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		From:      from,
		To:        to,
		Type:      msgType,
		Payload:   Payload{Text: text},
	}
}

func NewUserMessage(to, text string) *Message {
	return NewMessage(Human, to, TypeMessage, text)
}

func NewAgentMessage(from, to, text string) *Message {
	return NewMessage(from, to, TypeMessage, text)
}

func (m *Message) WithMetadata(key, value string) *Message {
	if m.Payload.Metadata == nil {
		m.Payload.Metadata = make(map[string]string)
	}
	m.Payload.Metadata[key] = value
	return m
}

func (m *Message) WithContext(sessionID string, turnNumber int) *Message {
	m.Context = &Context{SessionID: sessionID, TurnNumber: turnNumber}
	return m
}

func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func FromJSON(data []byte) (*Message, error) {
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
