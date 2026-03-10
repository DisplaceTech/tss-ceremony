#!/bin/bash

# verify-determinism.sh
# Verifies that running tss-ceremony with --fixed produces byte-for-byte
# identical output across multiple runs.
#
# Exit codes:
#   0  All outputs are identical — determinism verified.
#   1  Outputs differ, or a build/runtime error occurred.

set -euo pipefail

###############################################################################
# Helpers
###############################################################################

PASS="✓"
FAIL="✗"

info()  { echo "  $*"; }
ok()    { echo "${PASS} $*"; }
fail()  { echo "${FAIL} $*" >&2; }

###############################################################################
# Setup
###############################################################################

# Work from the project root regardless of where the script is invoked from.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${PROJECT_ROOT}"

BINARY="./tss-ceremony-test-$$"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}" "${BINARY}" 2>/dev/null || true' EXIT

echo "=== Determinism Verification for --fixed Mode ==="
echo ""

###############################################################################
# Build
###############################################################################

echo "Building tss-ceremony..."
go build -o "${BINARY}" .
ok "Build succeeded"
echo ""

###############################################################################
# Configuration: capture stderr+stdout for the non-interactive ceremony output.
# The TUI uses alt-screen and does not print to stdout when it runs; we rely on
# the post-TUI ceremony output printed after p.Run() returns.
# In CI / non-interactive environments the bubbletea program exits quickly, but
# to be safe we run with a flag that avoids the TUI entirely by redirecting
# stdin from /dev/null (bubbletea exits immediately when stdin is not a tty).
###############################################################################

RUNS=3          # number of independent runs to compare
RUN_OUTPUTS=()  # array of output file paths

echo "Running ceremony ${RUNS} times with --fixed flag..."
echo ""

for i in $(seq 1 ${RUNS}); do
    OUT_FILE="${TMPDIR}/run_${i}.txt"
    info "Run ${i}/${RUNS}..."

    # Run with --fixed. Redirect stdin from /dev/null so bubbletea detects a
    # non-tty environment and exits without needing user input.
    # Capture both stdout and stderr for a complete comparison.
    "${BINARY}" --fixed < /dev/null > "${OUT_FILE}" 2>&1 || true

    RUN_OUTPUTS+=("${OUT_FILE}")
    info "  Output captured to ${OUT_FILE} ($(wc -c < "${OUT_FILE}") bytes)"
done

echo ""

###############################################################################
# Compare
###############################################################################

echo "Comparing outputs..."
echo ""

REFERENCE="${RUN_OUTPUTS[0]}"
ALL_MATCH=true

for i in $(seq 2 ${RUNS}); do
    idx=$((i - 1))
    CANDIDATE="${RUN_OUTPUTS[${idx}]}"

    if cmp --silent "${REFERENCE}" "${CANDIDATE}"; then
        ok "Run 1 == Run ${i}: identical"
    else
        fail "Run 1 != Run ${i}: outputs differ!"
        echo ""
        echo "--- diff (run 1 vs run ${i}) ---"
        diff "${REFERENCE}" "${CANDIDATE}" || true
        echo "--- end diff ---"
        ALL_MATCH=false
    fi
done

echo ""

###############################################################################
# Validate that the fixed-mode output contains expected deterministic values
###############################################################################

echo "Validating fixed-mode output contents..."
echo ""

OUTPUT_TEXT="$(cat "${REFERENCE}")"

# The ceremony should always print these fields when running with --fixed.
REQUIRED_FIELDS=(
    "Fixed: true"
    "Phantom Public Key:"
    "Signature R:"
    "Signature S:"
    "Ceremony Complete"
)

for field in "${REQUIRED_FIELDS[@]}"; do
    if echo "${OUTPUT_TEXT}" | grep -q "${field}"; then
        ok "Output contains: ${field}"
    else
        fail "Output is missing: ${field}"
        ALL_MATCH=false
    fi
done

echo ""

###############################################################################
# Check that the fixed-mode signature values are stable (non-empty and constant)
###############################################################################

echo "Checking deterministic crypto values..."
echo ""

PHANTOM=$(echo "${OUTPUT_TEXT}" | grep "Phantom Public Key:" | awk '{print $NF}')
SIG_R=$(echo "${OUTPUT_TEXT}"   | grep "Signature R:"       | awk '{print $NF}')
SIG_S=$(echo "${OUTPUT_TEXT}"   | grep "Signature S:"       | awk '{print $NF}')

if [[ -n "${PHANTOM}" ]]; then
    ok "Phantom public key is present: ${PHANTOM:0:16}..."
else
    fail "Phantom public key is empty"
    ALL_MATCH=false
fi

if [[ -n "${SIG_R}" ]]; then
    ok "Signature R is present: ${SIG_R:0:16}..."
else
    fail "Signature R is empty"
    ALL_MATCH=false
fi

if [[ -n "${SIG_S}" ]]; then
    ok "Signature S is present: ${SIG_S:0:16}..."
else
    fail "Signature S is empty"
    ALL_MATCH=false
fi

# Verify the fixed-mode values match across all individual run outputs
for i in $(seq 2 ${RUNS}); do
    idx=$((i - 1))
    RUN_TEXT="$(cat "${RUN_OUTPUTS[${idx}]}")"

    PHANTOM_I=$(echo "${RUN_TEXT}" | grep "Phantom Public Key:" | awk '{print $NF}')
    SIG_R_I=$(echo "${RUN_TEXT}"   | grep "Signature R:"       | awk '{print $NF}')
    SIG_S_I=$(echo "${RUN_TEXT}"   | grep "Signature S:"       | awk '{print $NF}')

    if [[ "${PHANTOM_I}" == "${PHANTOM}" ]]; then
        ok "Run ${i}: Phantom public key matches reference"
    else
        fail "Run ${i}: Phantom public key differs (got ${PHANTOM_I:0:16}..., expected ${PHANTOM:0:16}...)"
        ALL_MATCH=false
    fi

    if [[ "${SIG_R_I}" == "${SIG_R}" ]]; then
        ok "Run ${i}: Signature R matches reference"
    else
        fail "Run ${i}: Signature R differs (got ${SIG_R_I:0:16}..., expected ${SIG_R:0:16}...)"
        ALL_MATCH=false
    fi

    if [[ "${SIG_S_I}" == "${SIG_S}" ]]; then
        ok "Run ${i}: Signature S matches reference"
    else
        fail "Run ${i}: Signature S differs (got ${SIG_S_I:0:16}..., expected ${SIG_S:0:16}...)"
        ALL_MATCH=false
    fi
done

echo ""

###############################################################################
# Summary
###############################################################################

if ${ALL_MATCH}; then
    echo "=== RESULT: PASSED — Fixed mode is fully deterministic ==="
    exit 0
else
    echo "=== RESULT: FAILED — Fixed mode is NOT deterministic ===" >&2
    exit 1
fi
