# TSS Ceremony

Interactive TUI for DKLS23 2-of-2 threshold ECDSA.

## Overview

This project provides a terminal user interface for conducting threshold signature scheme (TSS) ceremonies using the DKLS23 protocol for 2-of-2 threshold ECDSA key generation.

## Setup

### Prerequisites

- Go 1.25.7 or later
- Git (for version control and hooks)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/DisplaceTech/tss-ceremony.git
   cd tss-ceremony
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   go build ./...
   ```

### Running

```bash
./tss-ceremony [options]
```

#### Options

- `--fixed` - Use fixed seed for deterministic run

## Development

### Running Tests

```bash
go test ./...
```

### Code Quality

```bash
go vet ./...
```

## Contribution Guidelines

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linting (`go test ./...` and `go vet ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Code Style

- Follow Go formatting conventions (`gofmt`)
- Write tests for new functionality
- Keep commits atomic and well-documented

## License

This project is licensed under the terms found in the [LICENSE](LICENSE) file.

## Security

This project handles cryptographic operations. Please report any security vulnerabilities responsibly.
