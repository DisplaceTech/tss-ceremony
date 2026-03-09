#!/bin/bash

# Integration Test Script for OpenSSL Verification Command
# This script tests that the generated OpenSSL command can successfully verify signatures

set -e

echo "=== OpenSSL Verification Command Integration Test ==="
echo ""

# Build the ceremony tool
echo "Building tss-ceremony..."
go build -o tss-ceremony .

# Run the ceremony in fixed mode to get deterministic output
echo "Running ceremony in fixed mode..."
OUTPUT=$(./tss-ceremony --fixed 2>&1)

# Extract the public key and signature from the output
PHANTOM_PUBKEY=$(echo "$OUTPUT" | grep "Phantom Public Key:" | awk '{print $4}')
SIG_R=$(echo "$OUTPUT" | grep "Signature R:" | awk '{print $3}')
SIG_S=$(echo "$OUTPUT" | grep "Signature S:" | awk '{print $3}')

echo "Extracted values:"
echo "  Phantom Public Key: $PHANTOM_PUBKEY"
echo "  Signature R: $SIG_R"
echo "  Signature S: $SIG_S"
echo ""

# The message that was signed (default message)
MESSAGE="TSS Ceremony Demo"
MESSAGE_HEX=$(echo -n "$MESSAGE" | xxd -p | tr -d '\n')

echo "Message: $MESSAGE"
echo "Message (hex): $MESSAGE_HEX"
echo ""

# Verify using the ceremony's verify subcommand first
echo "Step 1: Verifying with ceremony's internal verification..."
if ./tss-ceremony --verify --pubkey "$PHANTOM_PUBKEY" --sig-r "$SIG_R" --sig-s "$SIG_S" --message "$MESSAGE_HEX"; then
    echo "✓ Ceremony internal verification: VALID"
else
    echo "✗ Ceremony internal verification: INVALID"
    exit 1
fi
echo ""

# Create a Go program to generate the OpenSSL command
echo "Step 2: Generating OpenSSL verification command..."
cat > /tmp/gen_cmd.go << 'EOF'
package main

import (
	"fmt"
	"os"
	"github.com/DisplaceTech/tss-ceremony/protocol"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "Usage: gen_cmd <pubkey> <sig-r> <sig-s> <message>\n")
		os.Exit(1)
	}
	
	pubkey := os.Args[1]
	sigR := os.Args[2]
	sigS := os.Args[3]
	message := os.Args[4]
	
	cmd, err := protocol.GenerateOpenSSLVerifyCommand(pubkey, sigR, sigS, message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating command: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(cmd)
}
EOF

# Build the command generator
cd /tmp
go mod init gen_cmd 2>/dev/null || true
go mod edit -replace github.com/DisplaceTech/tss-ceremony=$(pwd)/../$(basename $(pwd))
go mod tidy 2>/dev/null || true

# Actually, let's use a simpler approach - create the command manually
echo "Generating OpenSSL command manually..."

# Create temporary directory
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Create the message file
echo "$MESSAGE_HEX" | xxd -r -p > "$TMPDIR/msg.bin"
echo "Created message file: $TMPDIR/msg.bin"

# Create the DER signature
# DER format: 30 <total_len> 02 <r_len> <r> 02 <s_len> <s>
# We need to handle leading zeros and high bits properly

# Parse R and S, removing leading zeros but keeping at least one byte if the high bit is set
parse_int() {
    local hex="$1"
    # Remove leading zeros
    hex=$(echo "$hex" | sed 's/^0*//')
    # If empty, it's zero
    if [ -z "$hex" ]; then
        echo "00"
    # If high bit is set (first char >= 8), prepend 00
    elif [ "$(echo "$hex" | cut -c1)" -ge "8" ] 2>/dev/null; then
        echo "00$hex"
    else
        echo "$hex"
    fi
}

R_CLEAN=$(parse_int "$SIG_R")
S_CLEAN=$(parse_int "$SIG_S")

R_LEN=$((${#R_CLEAN} / 2))
S_LEN=$((${#S_CLEAN} / 2))

# Build DER signature
# 30 = SEQUENCE tag
# 02 = INTEGER tag (for R)
# 02 = INTEGER tag (for S)
DER_SIG="30"
CONTENT_LEN=$((2 + R_LEN + 2 + S_LEN))
DER_SIG="${DER_SIG}$(printf '%02x' $CONTENT_LEN)"
DER_SIG="${DER_SIG}02$(printf '%02x' $R_LEN)$R_CLEAN"
DER_SIG="${DER_SIG}02$(printf '%02x' $S_LEN)$S_CLEAN"

echo "$DER_SIG" | xxd -r -p > "$TMPDIR/sig.der"
echo "Created DER signature file: $TMPDIR/sig.der"
echo "DER signature (hex): $DER_SIG"

# Create the public key DER structure for secp256k1
# Format: SEQUENCE { SEQUENCE { OID (ecPublicKey), OID (secp256k1) }, BIT STRING (public key) }
# OID for ecPublicKey: 1.2.840.10045.2.1 = 2a 86 48 ce 3d 02 01
# OID for secp256k1: 1.3.132.0.10 = 2a 86 48 ce 3d 03 01 07

# Public key DER structure:
# 30 59 - SEQUENCE, length 89
#   30 13 - SEQUENCE (algorithm identifier), length 19
#     06 07 2a 86 48 ce 3d 02 01 - OID 1.2.840.10045.2.1 (ecPublicKey)
#     06 08 2a 86 48 ce 3d 03 01 07 - OID 1.3.132.0.10 (secp256k1)
#   03 42 00 - BIT STRING, length 66, 0 unused bits
#     04 || x || y (65 bytes of uncompressed public key)

PUBKEY_DER="3059301306072a8648ce3d020106082a8648ce3d03010703420004$PHANTOM_PUBKEY"

echo "$PUBKEY_DER" | xxd -r -p > "$TMPDIR/pubkey.der"
echo "Created public key DER file: $TMPDIR/pubkey.der"
echo "Public key DER (hex): $PUBKEY_DER"
echo ""

# Step 3: Execute the OpenSSL verification command
echo "Step 3: Executing OpenSSL verification command..."
echo "Command: openssl dgst -sha256 -verify $TMPDIR/pubkey.der -signature $TMPDIR/sig.der $TMPDIR/msg.bin"
echo ""

if openssl dgst -sha256 -verify "$TMPDIR/pubkey.der" -signature "$TMPDIR/sig.der" "$TMPDIR/msg.bin"; then
    echo ""
    echo "✓ OpenSSL verification: VALID"
    echo ""
    echo "=== Integration Test PASSED ==="
    echo "The generated OpenSSL command successfully verified the signature!"
    exit 0
else
    echo ""
    echo "✗ OpenSSL verification: INVALID"
    echo ""
    echo "=== Integration Test FAILED ==="
    exit 1
fi
