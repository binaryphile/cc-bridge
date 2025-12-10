package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/binaryphile/cc-bridge/internal/schema"
)

type Queue struct {
	dir string
	mu  sync.RWMutex
}

type Manager struct {
	baseDir string
	queues  map[string]*Queue
	mu      sync.RWMutex
}

func NewManager(baseDir string) (*Manager, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create queue directory: %w", err)
	}
	return &Manager{
		baseDir: baseDir,
		queues:  make(map[string]*Queue),
	}, nil
}

func (m *Manager) GetQueue(agent string) (*Queue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if q, ok := m.queues[agent]; ok {
		return q, nil
	}

	queueDir := filepath.Join(m.baseDir, agent)
	if err := os.MkdirAll(queueDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create queue for %s: %w", agent, err)
	}

	q := &Queue{dir: queueDir}
	m.queues[agent] = q
	return q, nil
}

func (q *Queue) Enqueue(msg *schema.Message) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	filename := fmt.Sprintf("%d_%s.json", msg.Timestamp.UnixNano(), msg.ID)
	path := filepath.Join(q.dir, filename)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

func (q *Queue) Dequeue() (*schema.Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	files, err := q.listFiles()
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	path := filepath.Join(q.dir, files[0])
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	msg, err := schema.FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("failed to remove message file: %w", err)
	}
	return msg, nil
}

func (q *Queue) Peek() (*schema.Message, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	files, err := q.listFiles()
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	path := filepath.Join(q.dir, files[0])
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}
	return schema.FromJSON(data)
}

func (q *Queue) Len() (int, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	files, err := q.listFiles()
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

func (q *Queue) List() ([]*schema.Message, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	files, err := q.listFiles()
	if err != nil {
		return nil, err
	}

	messages := make([]*schema.Message, 0, len(files))
	for _, file := range files {
		path := filepath.Join(q.dir, file)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read message %s: %w", file, err)
		}
		msg, err := schema.FromJSON(data)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal message %s: %w", file, err)
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (q *Queue) Clear() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	files, err := q.listFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		path := filepath.Join(q.dir, file)
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", file, err)
		}
	}
	return nil
}

func (q *Queue) listFiles() ([]string, error) {
	entries, err := os.ReadDir(q.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read queue directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}
