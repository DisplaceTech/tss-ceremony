package tests

import (
	"os"
	"testing"
)

// TestMain runs all integration tests
func TestMain(m *testing.M) {
	// Run all tests
	code := m.Run()
	os.Exit(code)
}
