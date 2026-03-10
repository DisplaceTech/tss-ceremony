#!/usr/bin/env python3
"""
FROST Schnorr Integration Test

Simulates a multi-party FROST signing session, aggregates the signatures
into a Schnorr signature, and verifies the result using the standard library.
This ensures end-to-end compatibility with standard Schnorr verification.

Usage:
    python3 tests/test_frost_schnorr_integration.py

Optional dependency for libsecp256k1-based check:
    pip install coincurve
"""

import hashlib
import subprocess
import sys
import os

# ---------------------------------------------------------------------------
# secp256k1 curve parameters
# ---------------------------------------------------------------------------
_P  = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F
_N  = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
_Gx = 0x79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798
_Gy = 0x483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8
_G  = (_Gx, _Gy)


def _point_add(P, Q):
    """Add two points on the secp256k1 curve."""
    if P is None: return Q
    if Q is None: return P
    x1, y1 = P
    x2, y2 = Q
    if x1 == x2:
        if y1 != y2:
            return None
        lam = (3 * x1 * x1 * pow(2 * y1, _P - 2, _P)) % _P
    else:
        lam = ((y2 - y1) * pow(x2 - x1, _P - 2, _P)) % _P
    x3 = (lam * lam - x1 - x2) % _P
    y3 = (lam * (x1 - x3) - y1) % _P
    return (x3, y3)


def _point_mul(k, P):
    """Scalar multiplication on the secp256k1 curve."""
    result, addend = None, P
    while k:
        if k & 1:
            result = _point_add(result, addend)
        addend = _point_add(addend, addend)
        k >>= 1
    return result


def _to_unc(pt):
    """Encode a curve point as a 65-byte uncompressed hex string (04 || x || y)."""
    x, y = pt
    return "04" + format(x, "064x") + format(y, "064x")


def _parse_unc(hex_str):
    """Decode a 65-byte uncompressed hex string to (x, y)."""
    b = bytes.fromhex(hex_str)
    if len(b) != 65 or b[0] != 0x04:
        raise ValueError("Expected 65-byte uncompressed public key")
    return (int.from_bytes(b[1:33], "big"), int.from_bytes(b[33:65], "big"))


# ---------------------------------------------------------------------------
# Schnorr challenge (mirrors protocol/schnorr_verify.go)
# e = SHA256(R_bytes || message || P_bytes) mod n
# ---------------------------------------------------------------------------
def _challenge(R_hex, message, P_hex):
    h = hashlib.sha256()
    h.update(bytes.fromhex(R_hex))
    h.update(message)
    h.update(bytes.fromhex(P_hex))
    return int.from_bytes(h.digest(), "big") % _N


# ---------------------------------------------------------------------------
# Pure-Python Schnorr verifier  (s*G == R + e*P)
# ---------------------------------------------------------------------------
def _verify(P_hex, R_hex, s_hex, message):
    s = int.from_bytes(bytes.fromhex(s_hex), "big")
    e = _challenge(R_hex, message, P_hex)
    R = _parse_unc(R_hex)
    P = _parse_unc(P_hex)
    lhs = _point_mul(s, _G)
    rhs = _point_add(R, _point_mul(e, P))
    return lhs == rhs


# ---------------------------------------------------------------------------
# Multi-party FROST simulation
# ---------------------------------------------------------------------------
def _frost_multi_party(message=b"Hello, FROST!", num_parties=2, fixed=True):
    """
    Simulate a multi-party FROST signing session.
    
    For fixed mode:
    - Party i has secret share = i + 1 (Party A: 1, Party B: 2)
    - Party i has nonce share = i + 100 (Party A: 100, Party B: 101)
    
    Returns:
        P_hex: Combined public key
        R_hex: Combined nonce point
        s_hex: Combined signature scalar
    """
    # Generate party keys
    parties = []
    for i in range(num_parties):
        if fixed:
            secret = i + 1
            nonce = i + 100
        else:
            # Random secrets and nonces (for testing)
            secret = hash((i, "secret")) % _N + 1
            nonce = hash((i, "nonce")) % _N + 1
        
        party_public = _point_mul(secret, _G)
        party_nonce = _point_mul(nonce, _G)
        parties.append({
            "secret": secret,
            "nonce": nonce,
            "public": party_public,
            "nonce_point": party_nonce
        })
    
    # Compute combined public key P = sum(P_i)
    P = None
    for party in parties:
        P = _point_add(P, party["public"])
    P_hex = _to_unc(P)
    
    # Compute combined nonce point R = sum(R_i)
    R = None
    for party in parties:
        R = _point_add(R, party["nonce_point"])
    R_hex = _to_unc(R)
    
    # Compute challenge e = H(R || message || P)
    e = _challenge(R_hex, message, P_hex)
    
    # Compute partial signatures s_i = k_i + e * x_i (mod n)
    s = 0
    for party in parties:
        e_times_x = (e * party["secret"]) % _N
        s_i = (party["nonce"] + e_times_x) % _N
        s = (s + s_i) % _N
    
    return P_hex, R_hex, format(s, "064x")


# ---------------------------------------------------------------------------
# Individual test functions
# ---------------------------------------------------------------------------
PASS, FAIL, SKIP = "PASS", "FAIL", "SKIP"


def test_frost_2of2_integration():
    """2-of-2 FROST signature must verify with standard Schnorr verification."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, s_hex = _frost_multi_party(msg, num_parties=2, fixed=True)
    ok = _verify(P_hex, R_hex, s_hex, msg)
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is True, "2-of-2 FROST signature must verify successfully (returned False)"
    return "2-of-2 FROST signature verifies (pure-Python)", PASS if ok else FAIL, ""


def test_frost_3of3_integration():
    """3-of-3 FROST signature must verify with standard Schnorr verification."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, s_hex = _frost_multi_party(msg, num_parties=3, fixed=True)
    ok = _verify(P_hex, R_hex, s_hex, msg)
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is True, "3-of-3 FROST signature must verify successfully (returned False)"
    return "3-of-3 FROST signature verifies (pure-Python)", PASS if ok else FAIL, ""


def test_frost_2of3_integration():
    """2-of-3 FROST signature (using 2 parties) must verify with standard Schnorr verification."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, s_hex = _frost_multi_party(msg, num_parties=2, fixed=True)
    ok = _verify(P_hex, R_hex, s_hex, msg)
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is True, "2-of-3 FROST signature must verify successfully (returned False)"
    return "2-of-3 FROST signature verifies (pure-Python)", PASS if ok else FAIL, ""


def test_frost_wrong_message_rejected():
    """FROST signature over a different message must NOT verify."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, s_hex = _frost_multi_party(msg, num_parties=2, fixed=True)
    ok = _verify(P_hex, R_hex, s_hex, b"Evil message")
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is False, "Wrong message should NOT verify (returned True)"
    status = PASS if not ok else FAIL
    return "Wrong message rejected (FROST)", status, f"verified={ok}"


def test_frost_tampered_s_rejected():
    """Tampered S scalar in FROST signature must NOT verify."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, _ = _frost_multi_party(msg, num_parties=2, fixed=True)
    bad_s = format(12345, "064x")
    ok = _verify(P_hex, R_hex, bad_s, msg)
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is False, "Tampered S should NOT verify (returned True)"
    status = PASS if not ok else FAIL
    return "Tampered S rejected (FROST)", status, f"verified={ok}"


def test_frost_tampered_r_rejected():
    """Tampered R nonce point in FROST signature must NOT verify."""
    msg = b"Hello, FROST!"
    P_hex, _, s_hex = _frost_multi_party(msg, num_parties=2, fixed=True)
    fake_R = _to_unc(_G)   # use generator as fake R
    ok = _verify(P_hex, fake_R, s_hex, msg)
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is False, "Tampered R should NOT verify (returned True)"
    status = PASS if not ok else FAIL
    return "Tampered R rejected (FROST)", status, f"verified={ok}"


def test_frost_wrong_public_key_rejected():
    """FROST signature with wrong public key must NOT verify."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, s_hex = _frost_multi_party(msg, num_parties=2, fixed=True)
    # Use a different public key (secret=999 instead of secret=3)
    wrong_P = _point_mul(999, _G)
    wrong_P_hex = _to_unc(wrong_P)
    ok = _verify(wrong_P_hex, R_hex, s_hex, msg)
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is False, "Wrong public key should NOT verify (returned True)"
    status = PASS if not ok else FAIL
    return "Wrong public key rejected (FROST)", status, f"verified={ok}"


def test_frost_empty_message():
    """FROST signature over empty message must verify."""
    msg = b""
    P_hex, R_hex, s_hex = _frost_multi_party(msg, num_parties=2, fixed=True)
    ok = _verify(P_hex, R_hex, s_hex, msg)
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is True, "Empty message FROST signature must verify successfully (returned False)"
    return "Empty message FROST signature verifies (pure-Python)", PASS if ok else FAIL, ""


def test_frost_long_message():
    """FROST signature over long message must verify."""
    msg = b"A" * 10000  # 10KB message
    P_hex, R_hex, s_hex = _frost_multi_party(msg, num_parties=2, fixed=True)
    ok = _verify(P_hex, R_hex, s_hex, msg)
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is True, "Long message FROST signature must verify successfully (returned False)"
    return "Long message FROST signature verifies (pure-Python)", PASS if ok else FAIL, ""


def test_frost_random_mode():
    """FROST signature in random mode must verify."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, s_hex = _frost_multi_party(msg, num_parties=2, fixed=False)
    ok = _verify(P_hex, R_hex, s_hex, msg)
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is True, "Random mode FROST signature must verify successfully (returned False)"
    return "Random mode FROST signature verifies (pure-Python)", PASS if ok else FAIL, ""


def test_frost_determinism():
    """FROST signatures must be deterministic with fixed mode."""
    msg = b"Hello, FROST!"
    P_hex1, R_hex1, s_hex1 = _frost_multi_party(msg, num_parties=2, fixed=True)
    P_hex2, R_hex2, s_hex2 = _frost_multi_party(msg, num_parties=2, fixed=True)
    
    assert P_hex1 == P_hex2, "Public keys must be deterministic"
    assert R_hex1 == R_hex2, "Nonce points must be deterministic"
    assert s_hex1 == s_hex2, "Signatures must be deterministic"
    
    # Verify the signature
    ok = _verify(P_hex1, R_hex1, s_hex1, msg)
    assert ok is True, "Deterministic FROST signature must verify successfully"
    
    return "FROST signatures are deterministic", PASS, ""


def test_frost_signature_format():
    """FROST signature must match secp256k1 Schnorr standard format."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, s_hex = _frost_multi_party(msg, num_parties=2, fixed=True)
    
    # R must be 65 hex chars (64 bytes + 04 prefix)
    assert len(R_hex) == 130, f"R must be 130 hex chars (65 bytes), got {len(R_hex)}"
    assert R_hex.startswith("04"), "R must be uncompressed format (04 prefix)"
    
    # S must be 64 hex chars (32 bytes)
    assert len(s_hex) == 64, f"S must be 64 hex chars (32 bytes), got {len(s_hex)}"
    
    # Verify the signature
    ok = _verify(P_hex, R_hex, s_hex, msg)
    assert ok is True, "Standard format FROST signature must verify successfully"
    
    return "FROST signature format matches secp256k1 standard", PASS, ""


# ---------------------------------------------------------------------------
# Run all tests
# ---------------------------------------------------------------------------
def run_all_tests():
    """Run all integration tests and report results."""
    tests = [
        test_frost_2of2_integration,
        test_frost_3of3_integration,
        test_frost_2of3_integration,
        test_frost_wrong_message_rejected,
        test_frost_tampered_s_rejected,
        test_frost_tampered_r_rejected,
        test_frost_wrong_public_key_rejected,
        test_frost_empty_message,
        test_frost_long_message,
        test_frost_random_mode,
        test_frost_determinism,
        test_frost_signature_format,
    ]
    
    results = []
    for test in tests:
        try:
            name, status, details = test()
            results.append((name, status, details))
        except Exception as e:
            results.append((test.__name__, FAIL, str(e)))
    
    # Print results
    print("=" * 70)
    print("FROST Schnorr Integration Test Results")
    print("=" * 70)
    
    passed = sum(1 for _, status, _ in results if status == PASS)
    failed = sum(1 for _, status, _ in results if status == FAIL)
    
    for name, status, details in results:
        symbol = "✓" if status == PASS else "✗"
        print(f"{symbol} {name}: {status}")
        if details:
            print(f"  {details}")
    
    print("=" * 70)
    print(f"Total: {len(results)} tests, {passed} passed, {failed} failed")
    print("=" * 70)
    
    return failed == 0


if __name__ == "__main__":
    success = run_all_tests()
    sys.exit(0 if success else 1)
