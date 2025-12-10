# cc-bridge

Inter-instance communication channel for Claude Code instances.

## Overview

cc-bridge enables two Claude Code instances to exchange messages at human conversation speed via a turn-based broker with session resume.

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

## Installation

```bash
go build -o cc-bridge ./cmd/cc-bridge
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
./cc-bridge send --to agent-a "Hello, Agent A!"
./cc-bridge send --to agent-b "Hello, Agent B!"
```

### Inject messages (masquerade as another agent)

```bash
# Pretend to be Agent A sending to Agent B
./cc-bridge inject --as agent-a --to agent-b "I am agent-a speaking"

# Pretend to be Agent B sending to Agent A
./cc-bridge inject --as agent-b --to agent-a "I am agent-b speaking"
```

### Check status

```bash
./cc-bridge status
```

## Data Storage

- **Queues:** `<data-dir>/queues/<agent>/*.json`
- **Sessions:** `<data-dir>/sessions/sessions.json`

Default data directory: `~/.cc-bridge`

## Testing

```bash
# Run all tests (fast, uses mocks)
go test ./... -short

# Run with real Claude CLI (slower, requires claude binary)
go test ./... -v

# Run just e2e tests
go test ./internal/broker/... -run TestE2E -v
```

## Test Results

```
ok  github.com/tedlilley/cc-bridge/cmd/cc-bridge     0.003s
ok  github.com/tedlilley/cc-bridge/internal/broker   25.314s
ok  github.com/tedlilley/cc-bridge/internal/queue    0.009s
ok  github.com/tedlilley/cc-bridge/internal/schema   0.002s
ok  github.com/tedlilley/cc-bridge/internal/session  0.003s
```

## E2E Test Evidence

```
=== RUN   TestE2E_TwoAgentConversation
    Step 1: Sending secret to Agent A
    Agent A response: STORED
    Step 2: Asking Agent A to recall the secret
    Agent A recall response: DELTA-7
    Step 3: Testing Agent B independently
    Agent B response: PONG
    Agent A session: f3328d74-... (turns: 2)
    Agent B session: 2c87a5a2-... (turns: 1)
--- PASS: TestE2E_TwoAgentConversation (18.38s)

=== RUN   TestE2E_Injection
    Injecting message as Agent A to Agent B
    Agent B response to injected message: RECEIVED FROM A
--- PASS: TestE2E_Injection (6.93s)
```

## License

MIT
