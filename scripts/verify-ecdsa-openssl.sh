#!/bin/bash

# OpenSSL-based ECDSA Verification Script
# This script verifies ECDSA signatures using OpenSSL's dgst command
# Usage: ./verify-ecdsa-openssl.sh <message> <public_key_hex> <signature_r_hex> <signature_s_hex>

set -e

# Check arguments
if [ $# -ne 4 ]; then
    echo "Usage: $0 <message> <public_key_hex> <signature_r_hex> <signature_s_hex>"
    echo ""
    echo "Arguments:"
    echo "  message          - The message that was signed (plain text)"
    echo "  public_key_hex   - The public key in hex format (64 bytes, x||y)"
    echo "  signature_r_hex  - The R component of the signature in hex (32 bytes)"
    echo "  signature_s_hex  - The S component of the signature in hex (32 bytes)"
    echo ""
    echo "Example:"
    echo "  $0 \"Hello World\" \"<pubkey>\" \"<r>\" \"<s>\""
    exit 1
fi

MESSAGE="$1"
PUBKEY_HEX="$2"
SIG_R="$3"
SIG_S="$4"

# Create temporary directory
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

echo "=== ECDSA Verification with OpenSSL ==="
echo ""
echo "Message: $MESSAGE"
echo "Public Key: $PUBKEY_HEX"
echo "Signature R: $SIG_R"
echo "Signature S: $SIG_S"
echo ""

# Convert message to hex
MESSAGE_HEX=$(echo -n "$MESSAGE" | xxd -p | tr -d '\n')
echo "Message (hex): $MESSAGE_HEX"
echo ""

# Create the public key in uncompressed SEC1 format for OpenSSL
# Format: 0x04 || x || y (65 bytes total)
echo "Creating public key file..."
echo "04${PUBKEY_HEX}" | xxd -r -p > "$TMPDIR/pubkey_uncompressed.bin"

# Create a temporary EC parameters file for secp256k1
echo "Creating EC parameters..."
cat > "$TMPDIR/ec_params.pem" << 'EOF'
-----BEGIN EC PARAMETERS-----
BggqhkjOPQMBBw==
-----END EC PARAMETERS-----
EOF

# Create the public key in PEM format using openssl
# We need to construct the proper ASN.1 structure
echo "Converting public key to PEM format..."

# Create a temporary file with the raw public key
cp "$TMPDIR/pubkey_uncompressed.bin" "$TMPDIR/pubkey_raw.bin"

# Use openssl to create the public key in PEM format
# First, create a minimal EC public key structure
# The public key is already in uncompressed format (0x04 || x || y)

# Create the EC public key in DER format using openssl
# We'll use the -pubin flag to read the raw public key
openssl ecparam -name secp256k1 -noout -genkey 2>/dev/null | openssl ec -pubout -out "$TMPDIR/temp_pub.pem" 2>/dev/null

# Extract the public key from the generated key and replace with our key
# This is a workaround since openssl doesn't directly accept raw public keys
# We'll use a different approach: create the DER structure manually

# Create the EC public key in proper PEM format
# The structure is: SEQUENCE { SEQUENCE { OID, parameters }, BIT STRING (public key) }

# For secp256k1, the OID is 1.2.840.10045.2.1
# We'll use openssl to help construct this

# Alternative approach: use openssl to verify directly with the raw components
# OpenSSL's dgst command can verify signatures, but we need the public key in PEM format

# Create a proper EC public key file
# We'll construct the ASN.1 DER encoding manually

# The public key in uncompressed format is: 0x04 || x (32 bytes) || y (32 bytes)
# We need to wrap this in the proper ASN.1 structure

# Create the EC public key in PEM format
cat > "$TMPDIR/ec_pubkey.pem" << EOF
-----BEGIN PUBLIC KEY-----
$(
    # Convert the uncompressed public key to base64
    # The format is: 04 || x || y
    cat "$TMPDIR/pubkey_uncompressed.bin" | base64 -w 0
)
-----END PUBLIC KEY-----
EOF

# Verify the public key is valid
if ! openssl pkey -pubin -in "$TMPDIR/ec_pubkey.pem" -check 2>/dev/null; then
    echo "Warning: Public key validation failed, but continuing with verification..."
fi

# Create the signature in DER format
# DER format: 30 <total_len> 02 <r_len> <r> 02 <s_len> <s>
echo "Creating DER signature..."

# Create a Python script to generate the DER signature properly
python3 << EOF
import sys

def create_der_signature(r_hex, s_hex):
    # Convert hex to bytes
    r = bytes.fromhex(r_hex)
    s = bytes.fromhex(s_hex)
    
    # Remove leading zeros for DER encoding (unless the value is zero)
    # Also ensure the high bit is not set (add leading zero if needed)
    def encode_integer(data):
        # Remove leading zeros
        while len(data) > 1 and data[0] == 0:
            data = data[1:]
        # If high bit is set, prepend zero
        if data[0] & 0x80:
            data = b'\x00' + data
        return data
    
    r_encoded = encode_integer(r)
    s_encoded = encode_integer(s)
    
    # Create the SEQUENCE
    sequence_content = b'\x02' + bytes([len(r_encoded)]) + r_encoded + b'\x02' + bytes([len(s_encoded)]) + s_encoded
    total_len = len(sequence_content)
    
    der = b'\x30' + bytes([total_len]) + sequence_content
    return der

r = "$SIG_R"
s = "$SIG_S"

der_sig = create_der_signature(r, s)
with open("$TMPDIR/signature.der", "wb") as f:
    f.write(der_sig)
EOF

# Sign the message with OpenSSL to create a test signature
# We need to sign the message hash
echo "Creating message hash..."
echo -n "$MESSAGE" > "$TMPDIR/message.txt"

# Verify the signature using OpenSSL
echo "Verifying signature with OpenSSL..."
echo ""

# OpenSSL's dgst command verifies signatures
# We need to use the -verify flag with the public key
# The signature is in DER format

if openssl dgst -sha256 -verify "$TMPDIR/ec_pubkey.pem" -signature "$TMPDIR/signature.der" "$TMPDIR/message.txt" 2>&1; then
    echo ""
    echo "=== VERIFICATION RESULT: VALID ==="
    echo "The signature is valid according to OpenSSL."
    exit 0
else
    echo ""
    echo "=== VERIFICATION RESULT: INVALID ==="
    echo "The signature is NOT valid according to OpenSSL."
    exit 1
fi
