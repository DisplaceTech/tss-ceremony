# Retrospective: First Agentic Pass Review

## What Went Wrong

### 1. The Signing Ceremony Was Never Built (Scenes 5-11, 14)

The most critical gap: the entire signing ceremony TUI — the core value proposition of the project — was left as placeholder stubs. Scenes 0-4 (keygen) and scenes 15-19 (bonus FROST comparison) were implemented, but the middle of the ceremony (message hashing, nonce generation, OT, MtA, partial signatures, signature assembly, and the ceremony summary) was skipped entirely.

**Root cause:** The agents likely worked on independent milestones (M5 for keygen scenes, M9 for bonus scenes) without anyone tracking that M6 (signing scenes) and parts of M7 (scene 14) were never assigned or completed. The bonus scenes (lower priority P3) shipped before the core signing scenes (P0).

### 2. Scenes 12-13 Were Built But Never Wired

`verify.go` (Scene 12) and `impossibility.go` (Scene 13) were fully implemented with rich multi-step content, but `model.go` treated indices 5-14 as a single block of placeholders. The implementations existed on disk but were dead code — no user would ever see them.

**Root cause:** The agent that built scenes 12-13 didn't update `model.go` to wire them in. The agent that maintained `model.go` didn't check whether implementations existed before blanket-assigning placeholders. No integration test verified that all 20 scenes rendered non-placeholder content.

### 3. `ceremony.SignMessage()` Signed With the Wrong Key

The most architecturally significant bug: `SignMessage()` used `ecdsa.Sign()` with **Party A's private key alone**, producing a signature verifiable only against Party A's public key. This defeats the entire purpose of a threshold signing ceremony — the signature should verify against the combined (phantom) public key P = (a+b)·G.

The comment in the original code even said "This is a simplified signing for demonstration" — acknowledging the shortcut but never addressing it.

**Root cause:** The protocol layer was built bottom-up (individual functions first), and the orchestration in `Ceremony.SignMessage()` was written as a quick "get it compiling" stub that was never revisited. The tests were written against the stub behavior (verifying against Party A's key), so they passed and nobody noticed the conceptual error.

### 4. Message Handling Was Broken

The spec says `--message "Hello, threshold signatures!"` (plaintext), but `NewCeremony()` tried to hex-decode the message parameter. The default message was "TSS Ceremony Demo" instead of the spec's "Hello, threshold signatures!". This meant:
- `tss-ceremony --message "Hello world"` would crash with a hex decode error
- The default message didn't match the spec

**Root cause:** An early agent likely assumed the message was hex-encoded (common in crypto code) without reading the spec's CLI interface section.

### 5. No Data Flow From Protocol to TUI

The protocol layer computed real cryptographic values, but the TUI scenes had no access to them. The `scenes.Config` struct held only `FixedMode`, `Message`, `Speed`, and `NoColor` — no ceremony results. The TUI was effectively a slideshow disconnected from the actual crypto.

**Root cause:** The protocol and TUI were developed as isolated milestones (M1-M3 vs M4-M7) without a data contract between them. Nobody defined how ceremony results would flow into scene rendering.

### 6. `main.go` Had Post-TUI Demo Code

After the TUI exited, `main.go` ran the signing ceremony and then dumped protocol demo output (nonce generation, OT simulation, error handling demos) to the console. This was clearly scaffolding from early development that was never cleaned up. The signing happened _after_ the TUI, so the TUI could never display the results.

### 7. Scene Names Were Wrong

`model.go` labeled scenes 5-14 as "Placeholder" even though scenes 12-13 had real implementations. This made the header display misleading.

## What Went Right

- **Protocol layer is solid.** All cryptographic primitives (`GenerateNonceShare`, `ComputeNoncePublic`, `CombineNonces`, `MultiplicativeToAdditive`, `SimulateOT`, `ComputePartialSignature`, `CombinePartialSignatures`, `VerifyECDSASignature`) were correctly implemented and well-tested.
- **Test suite is comprehensive.** Table-driven tests, round-trip verification, OpenSSL compatibility testing, and edge case coverage were all done well.
- **Keygen scenes (0-4) look good.** Rolling hex animation, party coloring, phantom key display — all match the spec.
- **Bonus FROST scenes (15-19) are well-crafted.** Side-by-side comparisons, equation displays, and narrative text are educational and visually appealing.
- **Scene architecture is clean.** The `Scene` interface pattern with `Render()` and `Narrator()` methods works well and made adding new scenes straightforward.

## Recommendations for Strengthening the Agentic Harness

### 1. Add a Milestone Completion Gate

Before marking a milestone complete, require an automated check that verifies all deliverables are present. For this project:
- Each scene index should map to a non-placeholder implementation
- `go build ./...`, `go vet ./...`, and `go test ./...` must pass
- A smoke test should run the TUI with `--fixed` and verify it doesn't crash

### 2. Define Data Contracts Between Milestones

When milestones have producer-consumer relationships (protocol produces data, TUI consumes it), define the interface explicitly in the spec. For example:
```
M3 produces: CeremonyResult{SignatureR, SignatureS, NonceAHex, ...}
M6 consumes: CeremonyResult to display in scenes 5-11
```

This prevents the "two halves that don't connect" problem.

### 3. Add Integration Assertions to the Spec

The spec should include testable assertions like:
- "Scene 12 must display the actual signature R and S values from the ceremony"
- "The signature must verify against the combined public key, not an individual party's key"
- "`tss-ceremony --message 'test'` must not error"

These become acceptance tests that agents can run.

### 4. Prioritize Wiring Over Polish

Require agents to wire new components into the main entry point before polishing them. A rough but connected implementation is better than a polished but disconnected one. The rule: "If you build a scene, you must also update `model.go` to use it."

### 5. Track Placeholder Debt

Maintain a list of known placeholders/stubs. When an agent introduces a placeholder (like the `for i := 5; i < 15` loop), it should be flagged as blocking completion. A post-pass audit should verify zero placeholders remain.

### 6. Run End-to-End Tests

Add a test that creates a `Ceremony`, runs `SignMessage()`, builds a `Config` with `CeremonyData`, creates all 20 scenes via `NewModel()`, and verifies each scene's `Render()` output is non-empty and non-placeholder. This catches both "not wired" and "not implemented" bugs.

### 7. Use Spec-Derived Checklists

Convert the spec's milestone deliverables into a machine-readable checklist. After each agent completes work, diff the checklist against the codebase. Missing items become the next agent's assignment.
