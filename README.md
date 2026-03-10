# TSS Ceremony

Interactive Terminal User Interface (TUI) for demonstrating Threshold Signature Scheme (TSS) ceremonies, specifically DKLS23 2-of-2 threshold ECDSA signing and FROST Schnorr threshold signatures.

## Features

- **DKLS23 ECDSA Ceremony**: Animated demonstration of the DKLS23 threshold ECDSA signing protocol
- **FROST Schnorr Ceremony**: Animated demonstration of FROST (Flexible Round-Optimized Schnorr Threshold) threshold signatures
- **Educational TUI**: Step-by-step visualization of cryptographic protocol execution
- **Deterministic Testing**: Fixed-seed mode for reproducible testing and verification

## Architecture

- **`protocol/`**: Pure cryptographic logic (no TUI dependencies)
  - DKLS signing (ECDSA)
  - FROST signing (Schnorr)
  - Multiplicative-to-Additive (MTA) conversion
  - Oblivious Transfer (OT) implementation
  - Key generation and verification

- **`tui/`**: Bubble Tea terminal UI
  - Scene-based architecture
  - Animated protocol demonstrations
  - Real-time visualization of cryptographic operations

- **`tests/`**: Comprehensive test suite
  - Protocol correctness tests
  - Compatibility tests with standard libraries
  - Integration tests

## Installation

```bash
go install github.com/DisplaceTech/tss-ceremony@latest
```

## Usage

```bash
# Interactive mode
tss-ceremony

# Fixed-seed mode for deterministic testing
tss-ceremony --fixed

# Run specific scene
tss-ceremony --scene <scene_number>
```

## Testing

```bash
# Run all tests
go test ./...

# Run protocol tests
go test ./protocol/...

# Run TUI tests
go test ./tui/...

# Run compatibility tests
python3 tests/test_schnorr_compatibility.py
python3 tests/test_frost_schnorr_integration.py
```

## Schnorr Signature Compatibility

The FROST/Schnorr implementation produces signatures compatible with standard Schnorr verification tools. See [docs/compatibility.md](docs/compatibility.md) for detailed compatibility results.

## Dependencies

- **Go**: 1.25.7 or later
- **Bubble Tea**: github.com/charmbracelet/bubbletea
- **secp256k1**: github.com/decred/dcrd/dcrec/secp256k1/v4

## License

MIT License - see [LICENSE](LICENSE) for details.
