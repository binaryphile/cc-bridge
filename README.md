# cc-bridge

Inter-instance communication channel for Claude Code instances.

## Why?

Claude Code runs as an interactive CLI. Testing its behavior programmatically—like verifying session resume preserves context, or observing agent-to-agent interactions—requires infrastructure that doesn't exist out of the box.

cc-bridge provides:
- **Session management** - Automatic `--resume` handling across turns
- **Message routing** - Queue-based delivery between named agents
- **Human injection** - Send messages masquerading as any participant

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     CC-BRIDGE BROKER                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Session Mgr │  │ Message Q   │  │ Human Inject CLI    │  │
│  │ (A, B IDs)  │  │ A→B, B→A    │  │ inject --as A "msg" │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │ (invokes claude --resume)
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
  │ Claude A    │ │ Claude B    │ │ Human       │
  │ Session     │ │ Session     │ │ (inject)    │
  └─────────────┘ └─────────────┘ └─────────────┘
```

## Prerequisites

- **Go 1.21+**
- **Claude CLI** installed and authenticated (`claude --version` should work)

## Installation

```bash
go build -o cc-bridge ./cmd/cc-bridge
```

## Quick Start

```bash
# Run the E2E test to see it work (requires claude CLI)
go test ./internal/broker/... -run TestE2E -v
```

Expected output:
```
=== RUN   TestE2E_TwoAgentConversation
    Step 1: Sending secret to Agent A
    Agent A response: STORED
    Step 2: Asking Agent A to recall the secret
    Agent A recall response: DELTA-7
--- PASS: TestE2E_TwoAgentConversation (18.38s)
```

## Usage

### Start the broker

```bash
# Start with defaults (polls every 1s)
./cc-bridge start

# Custom data directory and poll interval
./cc-bridge start --data-dir /tmp/my-bridge --poll-interval 2s
```

### Send messages

```bash
# Send message as human to an agent
./cc-bridge send --to agent-a "Remember the code: DELTA-7"
# Agent A responds, broker logs response

./cc-bridge send --to agent-a "What was the code?"
# Agent A: "DELTA-7"
```

### Inject messages (masquerade as another agent)

```bash
# Pretend to be Agent A sending to Agent B
./cc-bridge inject --as agent-a --to agent-b "Hello from A"
# Agent B receives message appearing to be from Agent A
```

### Check status

```bash
./cc-bridge status
# Shows active sessions and queue depths
```

## Data Storage

- **Queues:** `<data-dir>/queues/<agent>/*.json`
- **Sessions:** `<data-dir>/sessions/sessions.json`

Default data directory: `~/.cc-bridge`

## Testing

```bash
# Fast unit tests (mocked, <1s)
go test ./... -short

# Full suite including E2E (~30s, requires claude CLI)
go test ./... -v

# Just E2E tests
go test ./internal/broker/... -run TestE2E -v
```

## Documentation

- [`docs/use-cases.md`](docs/use-cases.md) - Problem statement, scenarios, user stories
- [`docs/design.md`](docs/design.md) - Architecture, research findings, design decisions
- [`docs/implementation-log.md`](docs/implementation-log.md) - Build history, test counts

## License

MIT
