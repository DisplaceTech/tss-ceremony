// Package scenes holds data types shared between the protocol and TUI layers.
package scenes

// CeremonyData holds computed ceremony values for display.
// These are populated by the protocol layer before the TUI starts.
type CeremonyData struct {
	// Keygen
	PartyASecretHex string
	PartyBSecretHex string
	PartyAPubHex    string
	PartyBPubHex    string
	CombinedPubHex  string

	// Message
	MessageText string
	MessageHash string

	// Nonces
	NonceAHex       string
	NonceBHex       string
	NonceAPubHex    string
	NonceBPubHex    string
	CombinedRPubHex string
	RHex            string

	// OT
	OTInput0Hex string
	OTInput1Hex string
	OTChoiceBit int
	OTOutputHex string

	// MtA
	AlphaHex string
	BetaHex  string

	// Partial signatures
	PartialSigAHex string
	PartialSigBHex string

	// Final signature
	SignatureRHex string
	SignatureSHex string

	// Verification
	Valid          bool
	OpenSSLVerify  string
}

// Config holds the TUI configuration.
type Config struct {
	FixedMode bool
	Message   string
	Speed     string
	NoColor   bool
	Ceremony  *CeremonyData
}
