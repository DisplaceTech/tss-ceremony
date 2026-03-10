package protocol

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math/big"
	"sync"
)

// deterministicState holds the package-level deterministic RNG state.
// When fixed mode is active, all randomness in the protocol package is
// derived from a seeded counter-based hash construction, producing the
// same byte stream for the same seed across multiple runs.
var deterministicState = struct {
	mu      sync.Mutex
	active  bool
	seed    int64
	counter uint64
}{
	active:  false,
	seed:    0,
	counter: 0,
}

// SetDeterministic activates deterministic (fixed) mode for all cryptographic
// operations in the protocol package. The given seed is used to initialise
// the counter-based deterministic byte source.
//
// Call SetDeterministic(0) to reset to non-deterministic (OS entropy) mode.
func SetDeterministic(seed int64) {
	deterministicState.mu.Lock()
	defer deterministicState.mu.Unlock()
	if seed == 0 {
		deterministicState.active = false
		deterministicState.seed = 0
		deterministicState.counter = 0
	} else {
		deterministicState.active = true
		deterministicState.seed = seed
		deterministicState.counter = 0
	}
}

// IsDeterministic returns true if deterministic mode is currently active.
func IsDeterministic() bool {
	deterministicState.mu.Lock()
	defer deterministicState.mu.Unlock()
	return deterministicState.active
}

// deterministicReader implements io.Reader using a counter-mode hash construction.
// Each call to Read fills the buffer by hashing (seed || counter) with SHA-256,
// incrementing the counter as needed to fill arbitrary-length requests.
type deterministicReader struct {
	seed    int64
	counter *uint64
	mu      *sync.Mutex
	buf     []byte // leftover bytes from the last block
}

// Read fills p with deterministic pseudo-random bytes derived from the seed
// and a monotonically increasing counter. This satisfies io.Reader.
func (dr *deterministicReader) Read(p []byte) (int, error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	filled := 0
	for filled < len(p) {
		// Consume any buffered bytes first
		if len(dr.buf) > 0 {
			n := copy(p[filled:], dr.buf)
			dr.buf = dr.buf[n:]
			filled += n
			continue
		}

		// Produce a new 32-byte block: SHA-256(seed || counter)
		var seedBuf [8]byte
		var ctrBuf [8]byte
		binary.BigEndian.PutUint64(seedBuf[:], uint64(dr.seed))
		binary.BigEndian.PutUint64(ctrBuf[:], *dr.counter)
		*dr.counter++

		h := sha256.New()
		h.Write(seedBuf[:])
		h.Write(ctrBuf[:])
		dr.buf = h.Sum(nil) // 32 bytes
	}

	return len(p), nil
}

// randReader returns the appropriate io.Reader for random byte generation.
// In deterministic mode it returns a deterministic reader seeded from the
// package-level state; otherwise it returns crypto/rand.Reader.
func randReader() io.Reader {
	deterministicState.mu.Lock()
	active := deterministicState.active
	seed := deterministicState.seed
	counter := &deterministicState.counter
	mu := &deterministicState.mu
	deterministicState.mu.Unlock()

	if active {
		return &deterministicReader{seed: seed, counter: counter, mu: mu}
	}
	return rand.Reader
}

// randInt returns a uniform random integer in [0, max) using the protocol's
// random source (deterministic when fixed mode is active).
func randInt(max *big.Int) (*big.Int, error) {
	return randIntFromReader(randReader(), max)
}

// randIntFromReader generates a uniform random integer in [0, max) from the
// given reader using rejection sampling (same algorithm as crypto/rand.Int).
func randIntFromReader(reader io.Reader, max *big.Int) (*big.Int, error) {
	return rand.Int(reader, max)
}

// randBytes returns n pseudo-random bytes using the protocol's random source.
func randBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := randReader().Read(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
