package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("TSS Ceremony - Interactive TUI for DKLS23 2-of-2 threshold ECDSA")
	fmt.Println("Usage: tss-ceremony [options]")
	fmt.Println("Options:")
	fmt.Println("  --fixed    Use fixed seed for deterministic run")
	os.Exit(0)
}
