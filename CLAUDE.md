# tss-ceremony

Interactive TUI animating a DKLS23 2-of-2 threshold ECDSA signature ceremony.

## Critical Rules
- **Only delete files that are within your ticket's scope.** You may delete files listed in your task's `files_to_modify` if the task requires it (e.g., renaming, replacing). Never delete files outside your assigned scope — another agent created them and they are needed.
- **Do NOT overwrite go.mod or go.sum** unless you are adding a new dependency. If you need to add a dependency, use `go get`, not manual edits.
- **Preserve CLAUDE.md, .gitignore, LICENSE** — these are project-level files managed by the operator.

## Architecture
- `protocol/` — Pure crypto logic. No TUI dependencies. Fully testable.
  - DKLS signing (ECDSA) is the core path
  - FROST signing (Schnorr) is in `frost.go` for the bonus comparison scenes
- `tui/` — Bubbletea models and scenes. Each scene is a standalone component.
  - Scenes 0-14: Core DKLS ceremony
  - Scenes 15-19: Bonus Schnorr/FROST comparison (reuse animation components)
- `main.go` — CLI parsing, ceremony init, scene wiring.

## Key Decisions
- secp256k1 via `decred/dcrd/dcrec/secp256k1/v4` (same curve as Bitcoin/Ethereum)
- All crypto values are real — the ceremony produces a verifiable ECDSA signature
- Bonus FROST scenes use the same secp256k1 curve for apples-to-apples comparison
- Scenes are independent bubbletea models composed by tui/model.go
- The OT implementation is simplified for educational clarity, not production security

## Testing
- `go test ./protocol/...` — Verify crypto correctness
- `go test ./tui/...` — Verify scene transitions and view rendering
- `tss-ceremony --fixed` — Deterministic run for snapshot testing

## Style
- Party A: cyan. Party B: magenta. Shared: yellow. Phantom key: red.
- All hex values displayed in 8-char groups
- Narrator text in bottom panel, every scene

## Build Priority
- P0: Scenes 0-8, 12-13 (MVP)
- P1: Scenes 9, 14 (OT animation + security proof)
- P2: Scenes 10-11 (MtA + partial sigs detail)
- P3: Scenes 15-19 (Schnorr/FROST comparison bonus)
