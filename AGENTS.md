# Agents

- **Explore**: read-only codebase exploration and fast Q&A. Use to locate files, APIs, and architectural intent.
- **Build**: compiles backend and frontend, installs deps, produces artifacts (`frontend/dist/` and Go binary).
- **Test**: runs unit/integration/smoke tests and reports failures with concise summaries and failing output.
- **Fix**: small, focused source edits (lint, format, minor bugfixes, style). Prefer a single responsibility per PR.
- **Audit**: dependency checks, static analysis, license scanning, and security review.

## How to Use an Agent

- Purpose: state briefly what you need (goal, expected artifact, and scope).
- Inputs: list required files, env vars, or commands (e.g., `frontend` build requires `bun install`).
- Outputs: describe expected outputs (build artifacts, test report, patch, or PR).
- Ownership: assign a single owner for follow-up and questions.

Example invocation:

- `Explore` — find API handlers for `/api/folder` and return file paths.
- `Build` — install frontend deps, run `bun build`, then `go build` and return the binary path.

## Best Practices (how to make agents and their docs better)

- Be explicit: each agent entry should include Purpose, Inputs, Commands, Outputs, and Owner.
- Keep scope small: prefer many focused agents over one large, multi-purpose agent.
- Idempotency: design agent actions to be repeatable without side effects when possible.
- Timeouts & Limits: define reasonable timeouts, resource limits, and expected runtimes for long tasks.
- Logging & Observability: agents should emit clear logs, error causes, and steps so failures are actionable.
- Security boundaries: document any secrets, required permissions, and safe defaults (never print secrets).
- Test coverage: add automated tests or smoke checks for Build/Test agents to verify success criteria.
- CI integration: wire Build/Test to CI so agent actions map to reproducible pipelines (include sample commands).
- Versioning & Changelog: track agent behavior changes in repository changelogs or commit messages.
- Examples: include a short command snippet for common workflows to reduce friction for contributors.

## Improvements Checklist

- Add `Inputs` / `Outputs` blocks for each agent entry.
- Provide one-liner command examples for common flows (build, dev, test).
- Add ownership (GitHub handle or team) for each agent.
- Add CI job names that map to the agent (e.g., `ci/build`, `ci/test`).

Usage: pick an agent for the targeted task and follow the Inputs/Commands described above.
