# cc-bridge Design Document

## Research Findings

Before designing cc-bridge, we conducted empirical research on Claude Code's headless capabilities. The findings fundamentally shaped the architecture.

### Experiment Summary

| # | Experiment | Result | Key Finding |
|---|------------|--------|-------------|
| 1 | Basic stream-json output | SUCCESS | `--output-format stream-json` requires `--verbose` flag |
| 2 | stream-json input | SUCCESS | JSONL messages accepted via stdin with `--input-format stream-json` |
| 3 | Multi-turn batch | SUCCESS | Multiple stdin messages processed sequentially with context preserved |
| 4 | Async injection via named pipe | PARTIAL | First message works; Claude exits after turn completion |
| 5 | Session resume | SUCCESS | `--resume $SESSION_ID` preserves context across process invocations |

### Critical Insight: Turn-Based, Not Real-Time

**Original assumption:** Real-time bidirectional stdin/stdout pipes would enable persistent connections.

**Reality discovered:** Claude Code's headless mode is strictly **turn-based**. After processing a turn, the process exits. Mid-conversation injection is only possible during tool execution, not after turn completion.

This insight led to the turn-based broker architecture rather than a WebSocket-style persistent connection.

### Capability Matrix

| Capability | Supported | Notes |
|------------|-----------|-------|
| Batch multi-message processing | Yes | All messages must be provided upfront via stdin |
| Context preservation across turns | Yes | Within a session |
| Session resume by ID | Yes | Via `--resume` flag |
| JSON output parsing | Yes | Via `--output-format json` |
| Mid-processing guidance | Yes | While tools are running |
| Async injection after turn | No | Process exits after turn completion |
| Persistent connections | No | Each turn is a separate process |
| Real-time streaming input | No | Stdin is read once, then closed |

## Architecture

### System Overview

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

### Component Architecture

```
cc-bridge/
├── cmd/cc-bridge/          # CLI entrypoint
│   └── main.go             # Argument parsing, command dispatch
├── internal/
│   ├── schema/             # Message types and constants
│   │   └── message.go      # Message struct, agent IDs
│   ├── queue/              # File-based message queues
│   │   └── queue.go        # Queue + Manager
│   ├── session/            # Session persistence
│   │   └── session.go      # Session struct + Manager
│   └── broker/             # Core orchestration
│       ├── broker.go       # Broker, polling loop
│       └── executor.go     # Claude CLI wrapper
└── docs/                   # This documentation
```

### Message Flow

1. **Human/Agent sends message** → Enqueued to recipient's queue
2. **Broker polls queues** → Dequeues next message for each agent
3. **Broker invokes Claude CLI** → `claude --resume $SESSION -p "$MESSAGE" --output-format json --max-turns 1`
4. **Response captured** → Parsed from JSON output
5. **Session updated** → New session ID (if first turn), turn count incremented
6. **Response routed** → Enqueued to original sender's queue (for agent responses)

### Session Management

Each agent maintains:
- **Session ID** - UUID from Claude CLI, used for `--resume`
- **Turn Number** - Monotonically increasing count
- **Agent ID** - Stable identifier (e.g., "agent-a", "agent-b")

Sessions are persisted to disk, enabling broker restart without losing conversation context.

## Design Decisions

### Decision 1: File-Based Queues vs. Database

**Chosen:** File-based FIFO queues (one directory per agent, one JSON file per message)

**Rationale:**
- Zero external dependencies (no SQLite, Redis, etc.)
- Human-readable for debugging
- Atomic file operations for concurrency safety
- Sufficient for research/testing throughput

**Trade-off:** Not suitable for high-volume production use. Acceptable given non-goal of production deployment.

### Decision 2: Polling vs. Push

**Chosen:** Polling with configurable interval (default: 1 second)

**Rationale:**
- Claude CLI is process-based, not event-driven
- Polling matches "human conversation speed" requirement
- Simple implementation with predictable behavior
- No need for inter-process signaling complexity

**Trade-off:** Latency floor equals poll interval. Acceptable for human-speed interactions.

### Decision 3: JSON Message Schema

**Chosen:** Structured JSON with metadata

```json
{
  "id": "uuid",
  "timestamp": "2025-12-10T12:00:00Z",
  "from": "agent-a",
  "to": "agent-b",
  "type": "message",
  "payload": {
    "text": "Hello from Agent A"
  },
  "session_context": {
    "session_id": "uuid",
    "turn_number": 5
  }
}
```

**Rationale:**
- Self-documenting structure
- Extensible for future metadata (cost tracking, latency, etc.)
- Type field distinguishes user messages from injections
- Session context enables conversation reconstruction

### Decision 4: Executor Interface for Testability

**Chosen:** Abstract `Executor` interface wrapping Claude CLI

```go
type Executor interface {
    Execute(ctx context.Context, sessionID string, message string, isNew bool) (*ExecuteResult, error)
}
```

**Rationale:**
- Enables mock executor for fast unit tests
- Real executor for integration/e2e tests
- Clean separation between broker logic and CLI invocation
- Test suite runs in <1 second with mocks, ~30 seconds with real Claude

### Decision 5: Agent Flexibility

**Chosen:** Configurable agent list at broker initialization

**Rationale:**
- Initial hardcoded A/B was limiting
- Dynamic agent registration supports N-agent scenarios
- Agents method exposes registered agents for iteration

## References

Anthropic. (2025). *Claude Code overview*. https://code.claude.com/docs/en/overview

Anthropic. (2025). *CLI reference*. https://code.claude.com/docs/en/cli-reference

Anthropic. (2025). *Sub-agents*. https://code.claude.com/docs/en/sub-agents
