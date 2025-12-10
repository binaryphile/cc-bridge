---

2025-12-10T06:06:01Z | Plan: Claude Code Inter-Instance Communication Channel (cc-bridge)

# Plan: Claude Code Inter-Instance Communication Channel (cc-bridge)

**Date:** 2025-12-10
**Status:** Research complete, awaiting implementation approval

## Objective

Create a communication channel enabling:
1. Two Claude Code instances to exchange events at human conversation speed
2. Human operator injection as either agent
3. Comprehensive behavioral testing (tool usage, multi-turn dialogue, error handling)

## Research Findings Summary

### What Works (Experimentally Validated)
- `--output-format stream-json --verbose` - Structured JSONL output
- `--input-format stream-json` - Batch message injection via stdin
- Multi-turn batch processing - Context preserved across turns (SECRET=banana test confirmed)
- Session resume (`--resume $SESSION_ID`) - Documented, pending validation

### What Does NOT Work
- Async injection via named pipes - Claude exits when stdin exhausted
- Real-time bidirectional communication - No persistent connection model
- Keeping stdin open for continuous injection - Fails after first turn completes

### Architectural Pivot
- **Original assumption:** Real-time pipes between instances
- **Revised approach:** Turn-based broker with session resume

## Revised Architecture

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

## Implementation Components

1. **broker.go** - Session management, queue polling, Claude invocation
2. **queue.go** - File-based FIFO message queue
3. **cmd/inject/** - CLI for human message injection
4. **schema/** - Message type definitions
5. **README.md** - Usage documentation

## Success Criteria

- [ ] Message broker supporting pub-sub pattern
- [ ] JSON message schema for inter-agent communication
- [ ] Instance A publish mechanism (via session resume)
- [ ] Instance B subscribe mechanism (via polling)
- [ ] Human injection interface (`inject --as A --to B "msg"`)
- [ ] Documentation with usage examples
- [ ] Integration test demonstrating two-way communication

## Next Steps

1. **Immediate:** Clean up background processes, validate session resume
2. **Pending approval:** Implement cc-bridge broker in Go
3. **Future:** Test harness for behavioral testing at scale

## Artifacts Created

- `phase-1-contract.md` - Tandem Protocol contract (updated with research)
- `cc-bridge-experiments/FINDINGS.md` - Detailed research findings
- `testing-claude-code-behavior.md` - Updated guide (v2.4 → v2.5)

## References

- Claude Code Headless Mode Docs: https://code.claude.com/docs/en/headless
- Claude Code CLI Reference: https://code.claude.com/docs/en/cli-reference
- GitHub Issue #5703: Inter-Agent Communication Request
---

2025-12-10T06:51:45Z | Plan: cc-bridge Documentation (Phase 2)

# Plan: cc-bridge Documentation (Phase 2)

**Date:** 2025-12-10
**Status:** Planning (Step 1d - awaiting approval)

## Objective

Create three orthogonal documentation files for cc-bridge that capture all research, design decisions, and implementation history with proper APA citations.

## Documents to Create

| Document | Purpose | Contains | Excludes |
|----------|---------|----------|----------|
| `docs/use-cases.md` | Why this exists | Problems, scenarios, user stories | Architecture, code, dates |
| `docs/design.md` | How it works | Architecture, research, decisions, citations | Timeline, user stories |
| `docs/implementation-log.md` | What was built when | Chronological record, components, test counts | Rationale, scenarios |

## User Requirements

1. **Citations:** External sources only (not local scripts)
2. **Log granularity:** Component level (not individual TDD cycles)
3. **URLs:** Exact public URLs with `.md` suffix for LLM-friendly markdown versions

## URL Discovery

During research, discovered Claude Code docs have migrated:
- `docs.anthropic.com` → `docs.claude.com` → `code.claude.com`
- Markdown versions available by appending `.md` to URLs

**Verified base URL:** `https://code.claude.com/docs/en/`

## Key References (to be cited in APA format)

- Anthropic. (2025). *Claude Code CLI reference*. https://code.claude.com/docs/en/cli
- Anthropic. (2025). *Claude Code headless mode*. https://code.claude.com/docs/en/headless

## Plan Self-Assessment

**Grade:** B+ (86/100)

**Gaps identified:**
- No content outlines in contract (-4)
- Citation placement unspecified (-2)
- URL verification incomplete (-5)
- APA format details unconfirmed (-3)

**Improving plan to address these gaps before proceeding.**

## Context

This is Phase 2 of cc-bridge development. Phase 1 delivered the working implementation with 56 tests (3 e2e with real Claude). Phase 2 captures the research and decisions in permanent documentation.
---

2025-12-10T07:04:46Z | Progress: cc-bridge Phase 2 documentation complete (A grade)

# Progress: cc-bridge Phase 2 documentation complete (A grade)

## Progress

Created three orthogonal documentation files for cc-bridge:

1. **docs/use-cases.md** (112 lines) - Problem statement, 3 scenarios with concrete CLI examples, 4 user stories with acceptance criteria, non-goals section

2. **docs/design.md** (180 lines) - 5 experiment summaries, capability matrix, architecture diagram, message flow, 5 design decisions with rationale, 3 APA citations:
   - Anthropic. (2025). *Claude Code overview*. https://code.claude.com/docs/en/overview
   - Anthropic. (2025). *CLI reference*. https://code.claude.com/docs/en/cli-reference
   - Anthropic. (2025). *Sub-agents*. https://code.claude.com/docs/en/sub-agents

3. **docs/implementation-log.md** (215 lines) - 6 component phases, 56 tests documented with tables, E2E evidence with actual output, actual commit hashes (62043c8, a497a58)

**Improvements made to reach A grade:**
- Added actual git commit hashes to implementation-log.md
- Added concrete CLI examples to scenarios in use-cases.md
- Added context sections explaining the "why" for each scenario

**Orthogonality verified:** No content duplication across documents.

## Plan

1. Await user approval (Step 4b)
2. Commit docs/ to git
3. Delete phase-2-contract.md
4. Log contract to plan-log for archival
---

2025-12-10T07:06:59Z | Phase 2 Contract: cc-bridge Documentation

# Phase 2 Contract: cc-bridge Documentation

**Created:** 2025-12-10

## Step 1 Checklist
- [x] 1a: Presented understanding
- [x] 1b: Asked clarifying questions
- [x] 1b-answer: Received answers
- [x] 1c: Contract created (this file)
- [x] 1d: Approval received

## Objective

Create three orthogonal documentation files for cc-bridge that capture all research, design decisions, and implementation history with proper APA citations.

## User Requirements (from clarifying questions)

- **Citations:** External sources only (no local scripts)
- **Log granularity:** Component level (not individual TDD cycles)
- **URLs:** Exact public URLs with `.md` suffix for LLM-friendly markdown

## Success Criteria

- [x] `docs/use-cases.md` - Problem statement, scenarios, user stories (98 lines)
- [x] `docs/design.md` - Architecture, research findings, design decisions with APA citations (180 lines)
- [x] `docs/implementation-log.md` - Chronological component-level build record (218 lines)
- [x] All external references have APA-format citations (3 citations in design.md References section)
- [x] Documents are DRY and orthogonal (no duplication across files)

## Approach

1. Locate exact public URLs for Claude Code documentation
2. Write `use-cases.md` (what/why)
3. Write `design.md` (how/research, with citations)
4. Write `implementation-log.md` (when/what was built)

## Document Responsibilities (Orthogonality)

| Document | Purpose | Contains | Does NOT Contain |
|----------|---------|----------|------------------|
| use-cases.md | Why this exists | Problems, scenarios, user stories | Architecture, code, dates |
| design.md | How it works | Architecture, research, decisions, citations | Implementation timeline, user stories |
| implementation-log.md | What was built when | Chronological record, components, test counts | Architecture rationale, user scenarios |

## Verified URLs (for APA citations)

Base: `https://code.claude.com/docs/en/`

| Page | URL |
|------|-----|
| Overview | https://code.claude.com/docs/en/overview |
| CLI Reference | https://code.claude.com/docs/en/cli-reference |
| Interactive Mode | https://code.claude.com/docs/en/interactive-mode |
| Sub-agents | https://code.claude.com/docs/en/sub-agents |
| Hooks | https://code.claude.com/docs/en/hooks |
| Settings | https://code.claude.com/docs/en/settings |

Note: Markdown versions not available (404 on `.md` suffix). Will use standard HTML URLs.

## Content Outlines

### docs/use-cases.md
1. **Problem Statement** - Why inter-instance communication?
2. **Scenarios**
   - Testing Claude Code behavior programmatically
   - Agent A and Agent B collaboration
   - Human injection into agent conversations
3. **User Stories**
   - As a tester, I want to verify session resume works
   - As a developer, I want agents to pass messages at human speed
   - As a researcher, I want to inject prompts mid-conversation

### docs/design.md
1. **Research Findings**
   - Experiment 1-5 summaries with conclusions
   - Key insight: turn-based > real-time pipes
2. **Architecture**
   - Component diagram
   - Message flow
   - Session management
3. **Design Decisions**
   - File-based queues vs database
   - Polling vs push
   - JSON message schema
4. **References** (APA format)
   - Claude Code documentation citations

### docs/implementation-log.md
1. **Phase 1: Scaffolding** - Go module, directory structure
2. **Schema Component** - Message types (10 tests)
3. **Queue Component** - File-based FIFO (10 tests)
4. **Session Component** - Persistence (9 tests)
5. **Broker Component** - Core coordination (18 tests)
6. **CLI Component** - Commands (9 tests)
7. **E2E Validation** - Real Claude tests (evidence)

## Token Budget

Estimated: 15-25K tokens

---

## Actual Results

**Completed:** 2025-12-10

### Deliverables

| File | Lines | Purpose |
|------|-------|---------|
| `docs/use-cases.md` | 98 | Problem statement, 3 scenarios, 4 user stories, non-goals |
| `docs/design.md` | 180 | 5 experiments, architecture diagram, 5 design decisions, 3 APA citations |
| `docs/implementation-log.md` | 218 | 6 component phases, 56 tests documented, E2E evidence |
| **Total** | **496** | |

### Orthogonality Verification

| Content Type | use-cases | design | impl-log |
|--------------|-----------|--------|----------|
| Problem/Why | X | - | - |
| User stories | X | - | - |
| Experiments | - | X | - |
| Architecture | - | X | - |
| Design decisions | - | X | - |
| APA citations | - | X | - |
| Test counts | - | - | X |
| Component phases | - | - | X |
| E2E evidence | - | - | X |

No content duplication detected.

### Quality Verification

Spot-checked:
- `design.md` References section contains 3 APA-format citations
- `implementation-log.md` test counts match actual (56 total: 10+10+9+18+9)
- `use-cases.md` contains no architecture details or dates

### Self-Assessment

**Grade:** A (96/100)

**What went well:**
- Clean orthogonal separation achieved
- All research findings captured in design.md
- Accurate test counts with E2E evidence
- Actual commit hashes added (62043c8, a497a58)
- Concrete examples with real CLI commands added to scenarios

**Deductions:**
- `.md` suffix URLs not available (-3): Used HTML URLs instead (external constraint)
- Minor formatting polish possible (-1)

## Step 4 Checklist

- [x] 4a: Results presented to user
- [x] 4b: Approval received

## Approval

✅ APPROVED BY USER - 2025-12-10

Final deliverables: 3 documentation files (507 lines total)
