package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/tedlilley/cc-bridge/internal/broker"
	"github.com/tedlilley/cc-bridge/internal/queue"
	"github.com/tedlilley/cc-bridge/internal/schema"
	"github.com/tedlilley/cc-bridge/internal/session"
)

// Command represents a parsed CLI command
type Command struct {
	Command      string
	DataDir      string
	PollInterval time.Duration
	To           string
	As           string
	Message      string
}

// DefaultDataDir returns the default data directory
func DefaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/tmp", "cc-bridge")
	}
	return filepath.Join(home, ".cc-bridge")
}

// ParseArgs parses command line arguments
func ParseArgs(args []string) (*Command, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	cmd := &Command{
		Command:      args[0],
		DataDir:      DefaultDataDir(),
		PollInterval: time.Second,
	}

	validCommands := map[string]bool{
		"start":  true,
		"status": true,
		"send":   true,
		"inject": true,
	}

	if !validCommands[cmd.Command] {
		return nil, fmt.Errorf("invalid command: %s", cmd.Command)
	}

	// Parse flags based on command
	fs := flag.NewFlagSet(cmd.Command, flag.ContinueOnError)
	fs.StringVar(&cmd.DataDir, "data-dir", cmd.DataDir, "data directory")
	fs.DurationVar(&cmd.PollInterval, "poll-interval", cmd.PollInterval, "poll interval")
	fs.StringVar(&cmd.To, "to", "", "target agent")
	fs.StringVar(&cmd.As, "as", "", "agent to impersonate")

	if err := fs.Parse(args[1:]); err != nil {
		return nil, err
	}

	// Collect remaining args as message
	if fs.NArg() > 0 {
		cmd.Message = strings.Join(fs.Args(), " ")
	}

	return cmd, nil
}

func main() {
	cmd, err := ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Usage: cc-bridge <command> [options]\n")
		fmt.Fprintf(os.Stderr, "Commands: start, status, send, inject\n")
		os.Exit(1)
	}

	switch cmd.Command {
	case "start":
		runStart(cmd)
	case "status":
		runStatus(cmd)
	case "send":
		runSend(cmd)
	case "inject":
		runInject(cmd)
	}
}

func runStart(cmd *Command) {
	fmt.Printf("Starting cc-bridge broker...\n")
	fmt.Printf("Data directory: %s\n", cmd.DataDir)
	fmt.Printf("Poll interval: %v\n", cmd.PollInterval)

	// Initialize components
	qMgr, err := queue.NewManager(filepath.Join(cmd.DataDir, "queues"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create queue manager: %v\n", err)
		os.Exit(1)
	}

	sMgr, err := session.NewManager(filepath.Join(cmd.DataDir, "sessions"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create session manager: %v\n", err)
		os.Exit(1)
	}

	// Load existing sessions
	if err := sMgr.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load sessions: %v\n", err)
	}

	executor := broker.NewClaudeExecutor()
	b, err := broker.NewBroker(qMgr, sMgr, executor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create broker: %v\n", err)
		os.Exit(1)
	}

	// Initialize agents
	b.InitializeAgent(schema.AgentA)
	b.InitializeAgent(schema.AgentB)

	// Set response handler
	b.SetResponseHandler(func(msg *schema.Message) {
		fmt.Printf("[%s] %s -> %s: %s\n",
			msg.Timestamp.Format("15:04:05"),
			msg.From, msg.To, msg.Payload.Text)
	})

	// Handle shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		if err := sMgr.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save sessions: %v\n", err)
		}
		cancel()
	}()

	fmt.Println("Broker running. Press Ctrl+C to stop.")
	b.Run(ctx, cmd.PollInterval)
}

func runStatus(cmd *Command) {
	sMgr, err := session.NewManager(filepath.Join(cmd.DataDir, "sessions"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create session manager: %v\n", err)
		os.Exit(1)
	}

	if err := sMgr.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "No sessions found: %v\n", err)
		os.Exit(0)
	}

	sessions := sMgr.ListSessions()
	if len(sessions) == 0 {
		fmt.Println("No active sessions")
		return
	}

	fmt.Println("Active sessions:")
	for _, s := range sessions {
		status := "not started"
		if s.SessionID != "" {
			status = fmt.Sprintf("turn %d", s.TurnNumber)
		}
		fmt.Printf("  %s: %s\n", s.AgentID, status)
	}
}

func runSend(cmd *Command) {
	if cmd.To == "" {
		fmt.Fprintf(os.Stderr, "Error: --to is required\n")
		os.Exit(1)
	}
	if cmd.Message == "" {
		fmt.Fprintf(os.Stderr, "Error: message is required\n")
		os.Exit(1)
	}

	qMgr, err := queue.NewManager(filepath.Join(cmd.DataDir, "queues"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create queue manager: %v\n", err)
		os.Exit(1)
	}

	msg := schema.NewUserMessage(cmd.To, cmd.Message)
	q, err := qMgr.GetQueue(cmd.To)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get queue: %v\n", err)
		os.Exit(1)
	}

	if err := q.Enqueue(msg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Message sent to %s\n", cmd.To)
}

func runInject(cmd *Command) {
	if cmd.As == "" {
		fmt.Fprintf(os.Stderr, "Error: --as is required\n")
		os.Exit(1)
	}
	if cmd.To == "" {
		fmt.Fprintf(os.Stderr, "Error: --to is required\n")
		os.Exit(1)
	}
	if cmd.Message == "" {
		fmt.Fprintf(os.Stderr, "Error: message is required\n")
		os.Exit(1)
	}

	qMgr, err := queue.NewManager(filepath.Join(cmd.DataDir, "queues"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create queue manager: %v\n", err)
		os.Exit(1)
	}

	msg := schema.NewMessage(cmd.As, cmd.To, schema.TypeInject, cmd.Message)
	q, err := qMgr.GetQueue(cmd.To)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get queue: %v\n", err)
		os.Exit(1)
	}

	if err := q.Enqueue(msg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to inject message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Injected message as %s to %s\n", cmd.As, cmd.To)
}
