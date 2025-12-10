package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Session struct {
	AgentID    string    `json:"agent_id"`
	SessionID  string    `json:"session_id"`
	TurnNumber int       `json:"turn_number"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Manager struct {
	dir      string
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewManager(dir string) (*Manager, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}
	return &Manager{
		dir:      dir,
		sessions: make(map[string]*Session),
	}, nil
}

func (m *Manager) CreateSession(agentID string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	sess := &Session{
		AgentID:    agentID,
		SessionID:  "",
		TurnNumber: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	m.sessions[agentID] = sess
	return sess, nil
}

func (m *Manager) GetSession(agentID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, ok := m.sessions[agentID]
	if !ok {
		return nil, fmt.Errorf("session not found for agent: %s", agentID)
	}
	return sess, nil
}

func (m *Manager) SetSessionID(agentID, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[agentID]
	if !ok {
		return fmt.Errorf("session not found for agent: %s", agentID)
	}
	sess.SessionID = sessionID
	sess.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *Manager) IncrementTurn(agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[agentID]
	if !ok {
		return fmt.Errorf("session not found for agent: %s", agentID)
	}
	sess.TurnNumber++
	sess.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *Manager) ListSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}

	path := filepath.Join(m.dir, "sessions.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write sessions: %w", err)
	}
	return nil
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := filepath.Join(m.dir, "sessions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state to load
		}
		return fmt.Errorf("failed to read sessions: %w", err)
	}

	if err := json.Unmarshal(data, &m.sessions); err != nil {
		return fmt.Errorf("failed to unmarshal sessions: %w", err)
	}
	return nil
}
