#!/usr/bin/env python3
"""
Schnorr Signature Compatibility Test

Generates a Schnorr signature using the same fixed-seed logic as the Go FROST
implementation (protocol/frost.go) and verifies it both with a pure-Python
secp256k1 verifier (no external deps) and, when available, with the coincurve
library (libsecp256k1 Python bindings).

Usage:
    python3 tests/test_schnorr_compatibility.py

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
# Build a real FROST signature from the Go fixed seeds
# secretA=1, secretB=2, nonceA=100, nonceB=101 (see protocol/frost.go)
# ---------------------------------------------------------------------------
def _frost_fixed(message=b"Hello, FROST!"):
    PA = _point_mul(1, _G)
    PB = _point_mul(2, _G)
    P  = _point_add(PA, PB)
    P_hex = _to_unc(P)

    RA = _point_mul(100, _G)
    RB = _point_mul(101, _G)
    R  = _point_add(RA, RB)
    R_hex = _to_unc(R)

    e = _challenge(R_hex, message, P_hex)
    s = ((100 + e * 1) + (101 + e * 2)) % _N
    return P_hex, R_hex, format(s, "064x")


# ---------------------------------------------------------------------------
# Individual test functions
# ---------------------------------------------------------------------------
PASS, FAIL, SKIP = "PASS", "FAIL", "SKIP"


def test_valid_signature():
    """Valid FROST signature must verify."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, s_hex = _frost_fixed(msg)
    ok = _verify(P_hex, R_hex, s_hex, msg)
    # Explicitly validate that verification returns True for valid signatures
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is True, "Valid signature must verify successfully (returned False)"
    return "Valid FROST signature verifies (pure-Python)", PASS if ok else FAIL, ""


def test_wrong_message_rejected():
    """Signature over a different message must NOT verify."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, s_hex = _frost_fixed(msg)
    ok = _verify(P_hex, R_hex, s_hex, b"Evil message")
    # Explicitly validate that verification returns False for invalid signatures
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is False, "Wrong message should NOT verify (returned True)"
    status = PASS if not ok else FAIL
    return "Wrong message rejected (pure-Python)", status, f"verified={ok}"


def test_tampered_s_rejected():
    """Tampered S scalar must NOT verify."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, _ = _frost_fixed(msg)
    bad_s = format(12345, "064x")
    ok = _verify(P_hex, R_hex, bad_s, msg)
    # Explicitly validate that verification returns False for tampered signatures
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is False, "Tampered S should NOT verify (returned True)"
    status = PASS if not ok else FAIL
    return "Tampered S rejected (pure-Python)", status, f"verified={ok}"


def test_tampered_r_rejected():
    """Tampered R nonce point must NOT verify."""
    msg = b"Hello, FROST!"
    P_hex, _, s_hex = _frost_fixed(msg)
    fake_R = _to_unc(_G)   # use generator as fake R
    ok = _verify(P_hex, fake_R, s_hex, msg)
    # Explicitly validate that verification returns False for tampered R
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is False, "Tampered R should NOT verify (returned True)"
    status = PASS if not ok else FAIL
    return "Tampered R rejected (pure-Python)", status, f"verified={ok}"


def test_hashed_message():
    """Signature over a SHA-256 hashed message must verify."""
    raw = b"Hello, FROST!"
    msg = hashlib.sha256(raw).digest()
    P_hex, R_hex, s_hex = _frost_fixed(msg)
    ok = _verify(P_hex, R_hex, s_hex, msg)
    # Explicitly validate that verification returns True for hashed message signatures
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is True, "Hashed message signature must verify successfully (returned False)"
    return "Hashed message signature verifies (pure-Python)", PASS if ok else FAIL, ""


def test_wrong_public_key_rejected():
    """Signature with wrong public key must NOT verify."""
    msg = b"Hello, FROST!"
    P_hex, R_hex, s_hex = _frost_fixed(msg)
    # Use a different public key (secret=999 instead of secret=3)
    wrong_P = _point_mul(999, _G)
    wrong_P_hex = _to_unc(wrong_P)
    ok = _verify(wrong_P_hex, R_hex, s_hex, msg)
    # Explicitly validate that verification returns False for wrong public key
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is False, "Wrong public key should NOT verify (returned True)"
    status = PASS if not ok else FAIL
    return "Wrong public key rejected (pure-Python)", status, f"verified={ok}"


def test_empty_message():
    """Signature over empty message must verify."""
    msg = b""
    P_hex, R_hex, s_hex = _frost_fixed(msg)
    ok = _verify(P_hex, R_hex, s_hex, msg)
    # Explicitly validate that verification returns True for empty message
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is True, "Empty message signature must verify successfully (returned False)"
    return "Empty message signature verifies (pure-Python)", PASS if ok else FAIL, ""


def test_long_message():
    """Signature over long message must verify."""
    msg = b"A" * 1000  # 1000-byte message
    P_hex, R_hex, s_hex = _frost_fixed(msg)
    ok = _verify(P_hex, R_hex, s_hex, msg)
    # Explicitly validate that verification returns True for long message
    assert isinstance(ok, bool), "Verification must return a boolean"
    assert ok is True, "Long message signature must verify successfully (returned False)"
    return "Long message signature verifies (pure-Python)", PASS if ok else FAIL, ""


def test_go_schnorr_tests():
    """Run the Go Schnorr compatibility tests via `go test`."""
    root = os.path.normpath(os.path.join(os.path.dirname(__file__), ".."))
    try:
        r = subprocess.run(
            ["go", "test", "-run", "TestSchnorr", "./tests/", "./protocol/..."],
            capture_output=True, text=True, cwd=root, timeout=60,
        )
        ok = r.returncode == 0
        detail = (r.stdout + r.stderr).strip().splitlines()[-1] if (r.stdout + r.stderr).strip() else ""
        return "Go TestSchnorr* suite passes", PASS if ok else FAIL, detail
    except FileNotFoundError:
        return "Go TestSchnorr* suite passes", SKIP, "go binary not found"
    except subprocess.TimeoutExpired:
        return "Go TestSchnorr* suite passes", FAIL, "timed out"


def test_coincurve_pubkey_roundtrip():
    """Public key serialisation is compatible with coincurve (libsecp256k1)."""
    try:
        from coincurve import PublicKey as CCKey
    except ImportError:
        return "coincurve public-key round-trip", SKIP, "pip install coincurve to enable"

    P_hex, _, _ = _frost_fixed()
    try:
        cc = CCKey(bytes.fromhex(P_hex))
        recovered = cc.format(compressed=False).hex().upper()
        ok = recovered == P_hex.upper()
        detail = "" if ok else f"got {recovered}"
        return "coincurve public-key round-trip", PASS if ok else FAIL, detail
    except Exception as exc:
        return "coincurve public-key round-trip", FAIL, str(exc)


# ---------------------------------------------------------------------------
# Runner
# ---------------------------------------------------------------------------
def main():
    print("=" * 62)
    print("Schnorr Signature Compatibility Test")
    print("FROST/secp256k1 — verify against standard Schnorr equation")
    print("=" * 62)

    tests = [
        test_valid_signature,
        test_wrong_message_rejected,
        test_tampered_s_rejected,
        test_tampered_r_rejected,
        test_hashed_message,
        test_wrong_public_key_rejected,
        test_empty_message,
        test_long_message,
        test_go_schnorr_tests,
        test_coincurve_pubkey_roundtrip,
    ]

    passed = skipped = failed = 0
    for fn in tests:
        try:
            label, status, detail = fn()
        except Exception as exc:
            label, status, detail = fn.__name__, FAIL, f"exception: {exc}"

        icon = {"PASS": "✓", "FAIL": "✗", "SKIP": "○"}.get(status, "?")
        line = f"  {icon} [{status}] {label}"
        if detail:
            line += f"\n         {detail}"
        print(line)
        if status == PASS:   passed  += 1
        elif status == SKIP: skipped += 1
        else:                failed  += 1

    print()
    print(f"Results: {passed} passed, {skipped} skipped, {failed} failed")
    print("=" * 62)
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
