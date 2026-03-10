package protocol

import (
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// FrostAnimatedStep represents a single step in the FROST signing animation
type FrostAnimatedStep struct {
	Name        string
	Description string
	PartyAData  map[string]interface{}
	PartyBData  map[string]interface{}
	SharedData  map[string]interface{}
}

// FrostAnimatedScene represents the FROST signing animation at the protocol level
// It provides step-by-step progression through the FROST signing process
type FrostAnimatedScene struct {
	config      FROSTConfig
	signer      *FROSTSigner
	currentStep int
	steps       []FrostAnimatedStep
}

// NewFrostAnimatedScene creates a new FROST animated scene with the given configuration
func NewFrostAnimatedScene(config FROSTConfig) *FrostAnimatedScene {
	return &FrostAnimatedScene{
		config:      config,
		currentStep: 0,
	}
}

// Init initializes the FROST signer and prepares the animation steps
func (s *FrostAnimatedScene) Init() error {
	var err error
	s.signer, err = NewFROSTSigner(s.config)
	if err != nil {
		return fmt.Errorf("failed to initialize FROST signer: %w", err)
	}

	s.steps = s.createAnimationSteps()
	return nil
}

// createAnimationSteps creates the sequence of animation steps for FROST signing
func (s *FrostAnimatedScene) createAnimationSteps() []FrostAnimatedStep {
	return []FrostAnimatedStep{
		{
			Name:        "Key Generation (DKG)",
			Description: "Each party generates a secret share and broadcasts their public share",
			PartyAData:  make(map[string]interface{}),
			PartyBData:  make(map[string]interface{}),
			SharedData:  make(map[string]interface{}),
		},
		{
			Name:        "Nonce Generation",
			Description: "Each party generates a nonce share and broadcasts their nonce point",
			PartyAData:  make(map[string]interface{}),
			PartyBData:  make(map[string]interface{}),
			SharedData:  make(map[string]interface{}),
		},
		{
			Name:        "Compute Challenge",
			Description: "Both parties compute the challenge e = H(R, message, P)",
			PartyAData:  make(map[string]interface{}),
			PartyBData:  make(map[string]interface{}),
			SharedData:  make(map[string]interface{}),
		},
		{
			Name:        "Partial Signatures",
			Description: "Each party computes their partial signature s_i = k_i + e * x_i",
			PartyAData:  make(map[string]interface{}),
			PartyBData:  make(map[string]interface{}),
			SharedData:  make(map[string]interface{}),
		},
		{
			Name:        "Aggregation",
			Description: "Combine partial signatures: s = s_A + s_B",
			PartyAData:  make(map[string]interface{}),
			PartyBData:  make(map[string]interface{}),
			SharedData:  make(map[string]interface{}),
		},
		{
			Name:        "Verification",
			Description: "Verify the final signature using standard Schnorr verification",
			PartyAData:  make(map[string]interface{}),
			PartyBData:  make(map[string]interface{}),
			SharedData:  make(map[string]interface{}),
		},
	}
}

// Run executes the current step of the animation and returns the step data
func (s *FrostAnimatedScene) Run() (*FrostAnimatedStep, error) {
	if s.signer == nil {
		return nil, fmt.Errorf("scene not initialized, call Init() first")
	}

	if s.currentStep >= len(s.steps) {
		s.currentStep = 0 // Reset to beginning
	}

	step := &s.steps[s.currentStep]

	// Execute the step logic
	switch s.currentStep {
	case 0:
		// Key Generation step
		s.populateKeyGenStep(step)
	case 1:
		// Nonce Generation step
		if err := s.signer.GenerateNoncesWithFixed(s.config.Fixed); err != nil {
			return nil, fmt.Errorf("failed to generate nonces: %w", err)
		}
		s.populateNonceStep(step)
	case 2:
		// Compute Challenge step
		if err := s.signer.ComputeChallenge(s.config.Message); err != nil {
			return nil, fmt.Errorf("failed to compute challenge: %w", err)
		}
		s.populateChallengeStep(step)
	case 3:
		// Partial Signatures step
		if err := s.signer.ComputePartialSignatures(); err != nil {
			return nil, fmt.Errorf("failed to compute partial signatures: %w", err)
		}
		s.populatePartialSigsStep(step)
	case 4:
		// Aggregation step
		if err := s.signer.AggregateSignatures(); err != nil {
			return nil, fmt.Errorf("failed to aggregate signatures: %w", err)
		}
		s.populateAggregationStep(step)
	case 5:
		// Verification step
		s.populateVerificationStep(step)
	}

	s.currentStep++
	return step, nil
}

// populateKeyGenStep populates data for the key generation step
func (s *FrostAnimatedScene) populateKeyGenStep(step *FrostAnimatedStep) {
	if len(s.signer.Parties) > 0 {
		step.PartyAData["secret_share"] = s.signer.Parties[0].Secret.String()
		step.PartyAData["public_share"] = s.signer.Parties[0].Public.SerializeCompressed()
	}
	if len(s.signer.Parties) > 1 {
		step.PartyBData["secret_share"] = s.signer.Parties[1].Secret.String()
		step.PartyBData["public_share"] = s.signer.Parties[1].Public.SerializeCompressed()
	}
	if s.signer.P != nil {
		step.SharedData["combined_public_key"] = s.signer.P.SerializeCompressed()
	}
}

// KeyGen executes the key generation step of the FROST animation
// This step visualizes the generation of individual secret shares and public keys
// for each participant, then combines them into the shared public key.
func (s *FrostAnimatedScene) KeyGen() (*FrostAnimatedStep, error) {
	if s.signer == nil {
		return nil, fmt.Errorf("scene not initialized, call Init() first")
	}

	// Ensure we're at the key generation step
	if s.currentStep != 0 {
		s.currentStep = 0
	}

	step := &s.steps[s.currentStep]

	// Execute the key generation step
	s.populateKeyGenStep(step)

	// Advance to the next step
	s.currentStep++

	return step, nil
}

// NonceGen executes the nonce generation step of the FROST animation
// This step visualizes the generation of nonce shares and nonce points
// for each participant, then combines them into the shared nonce point R.
func (s *FrostAnimatedScene) NonceGen() (*FrostAnimatedStep, error) {
	if s.signer == nil {
		return nil, fmt.Errorf("scene not initialized, call Init() first")
	}

	// Ensure we're at the nonce generation step
	if s.currentStep != 1 {
		s.currentStep = 1
	}

	step := &s.steps[s.currentStep]

	// Execute the nonce generation step
	if err := s.signer.GenerateNoncesWithFixed(s.config.Fixed); err != nil {
		return nil, fmt.Errorf("failed to generate nonces: %w", err)
	}

	s.populateNonceStep(step)

	// Advance to the next step
	s.currentStep++

	return step, nil
}

// PartialSig executes the partial signature computation step of the FROST animation
// This step visualizes each participant computing their partial signature
// s_i = k_i + e * x_i (mod n), where k_i is the nonce share, e is the challenge,
// and x_i is the secret share.
func (s *FrostAnimatedScene) PartialSig() (*FrostAnimatedStep, error) {
	if s.signer == nil {
		return nil, fmt.Errorf("scene not initialized, call Init() first")
	}

	// Ensure we're at the partial signatures step
	if s.currentStep != 3 {
		s.currentStep = 3
	}

	step := &s.steps[s.currentStep]

	// Execute the partial signature computation step
	if err := s.signer.ComputePartialSignatures(); err != nil {
		return nil, fmt.Errorf("failed to compute partial signatures: %w", err)
	}

	s.populatePartialSigsStep(step)

	// Advance to the next step
	s.currentStep++

	return step, nil
}

// Aggregate executes the signature aggregation step of the FROST animation
// This step visualizes the combination of partial signatures from all participants
// into the final group signature s = s_A + s_B (mod n).
func (s *FrostAnimatedScene) Aggregate() (*FrostAnimatedStep, error) {
	if s.signer == nil {
		return nil, fmt.Errorf("scene not initialized, call Init() first")
	}

	// Ensure we're at the aggregation step
	if s.currentStep != 4 {
		s.currentStep = 4
	}

	step := &s.steps[s.currentStep]

	// Execute the signature aggregation step
	if err := s.signer.AggregateSignatures(); err != nil {
		return nil, fmt.Errorf("failed to aggregate signatures: %w", err)
	}

	s.populateAggregationStep(step)

	// Advance to the next step
	s.currentStep++

	return step, nil
}

// populateNonceStep populates data for the nonce generation step
func (s *FrostAnimatedScene) populateNonceStep(step *FrostAnimatedStep) {
	if len(s.signer.Parties) > 0 {
		step.PartyAData["nonce_share"] = s.signer.Parties[0].Nonce.String()
		step.PartyAData["nonce_point"] = s.signer.Parties[0].NoncePoint.SerializeCompressed()
	}
	if len(s.signer.Parties) > 1 {
		step.PartyBData["nonce_share"] = s.signer.Parties[1].Nonce.String()
		step.PartyBData["nonce_point"] = s.signer.Parties[1].NoncePoint.SerializeCompressed()
	}
	if s.signer.R != nil {
		step.SharedData["combined_nonce_point"] = s.signer.R.SerializeCompressed()
	}
}

// populateChallengeStep populates data for the challenge computation step
func (s *FrostAnimatedScene) populateChallengeStep(step *FrostAnimatedStep) {
	if s.signer.E != nil {
		step.SharedData["challenge"] = s.signer.E.String()
	}
	step.SharedData["message"] = string(s.config.Message)
	if s.signer.P != nil {
		step.SharedData["public_key"] = s.signer.P.SerializeCompressed()
	}
	if s.signer.R != nil {
		step.SharedData["nonce_point"] = s.signer.R.SerializeCompressed()
	}
}

// populatePartialSigsStep populates data for the partial signatures step
func (s *FrostAnimatedScene) populatePartialSigsStep(step *FrostAnimatedStep) {
	if len(s.signer.Parties) > 0 && s.signer.Parties[0].PartialSig != nil {
		step.PartyAData["partial_signature"] = s.signer.Parties[0].PartialSig.String()
	}
	if len(s.signer.Parties) > 1 && s.signer.Parties[1].PartialSig != nil {
		step.PartyBData["partial_signature"] = s.signer.Parties[1].PartialSig.String()
	}
}

// populateAggregationStep populates data for the aggregation step
func (s *FrostAnimatedScene) populateAggregationStep(step *FrostAnimatedStep) {
	if s.signer.S != nil {
		step.SharedData["final_signature_s"] = s.signer.S.String()
	}
	if s.signer.R != nil {
		step.SharedData["final_signature_r"] = s.signer.R.SerializeCompressed()
	}
}

// populateVerificationStep populates data for the verification step
func (s *FrostAnimatedScene) populateVerificationStep(step *FrostAnimatedStep) {
	if s.signer.S != nil {
		step.SharedData["signature_s"] = s.signer.S.String()
	}
	if s.signer.R != nil {
		step.SharedData["signature_r"] = s.signer.R.SerializeCompressed()
	}
	if s.signer.P != nil {
		step.SharedData["public_key"] = s.signer.P.SerializeCompressed()
	}
	step.SharedData["message"] = string(s.config.Message)
}

// Cleanup releases resources and resets the scene state
func (s *FrostAnimatedScene) Cleanup() {
	s.signer = nil
	s.currentStep = 0
	s.steps = nil
}

// GetCurrentStep returns the current step index
func (s *FrostAnimatedScene) GetCurrentStep() int {
	return s.currentStep
}

// GetTotalSteps returns the total number of steps
func (s *FrostAnimatedScene) GetTotalSteps() int {
	return len(s.steps)
}

// GetStepName returns the name of the current step
func (s *FrostAnimatedScene) GetStepName() string {
	if s.currentStep >= 0 && s.currentStep < len(s.steps) {
		return s.steps[s.currentStep].Name
	}
	return ""
}

// GetStepDescription returns the description of the current step
func (s *FrostAnimatedScene) GetStepDescription() string {
	if s.currentStep >= 0 && s.currentStep < len(s.steps) {
		return s.steps[s.currentStep].Description
	}
	return ""
}

// GetSignature returns the final FROST signature if available
func (s *FrostAnimatedScene) GetSignature() *FROSTSignature {
	if s.signer == nil || s.signer.S == nil || s.signer.R == nil {
		return nil
	}
	return &FROSTSignature{
		R: s.signer.R,
		S: s.signer.S,
	}
}

// GetCombinedPublicKey returns the combined public key
func (s *FrostAnimatedScene) GetCombinedPublicKey() *secp256k1.PublicKey {
	if s.signer == nil {
		return nil
	}
	return s.signer.P
}

// GetPartySecretShare returns a party's secret share (for testing/debugging)
func (s *FrostAnimatedScene) GetPartySecretShare(id int) *big.Int {
	if s.signer == nil || id < 0 || id >= len(s.signer.Parties) {
		return nil
	}
	return s.signer.Parties[id].Secret
}

// GetPartyPublicShare returns a party's public share
func (s *FrostAnimatedScene) GetPartyPublicShare(id int) *secp256k1.PublicKey {
	if s.signer == nil || id < 0 || id >= len(s.signer.Parties) {
		return nil
	}
	return s.signer.Parties[id].Public
}

// GetPartyNonceShare returns a party's nonce share
func (s *FrostAnimatedScene) GetPartyNonceShare(id int) *big.Int {
	if s.signer == nil || id < 0 || id >= len(s.signer.Parties) {
		return nil
	}
	return s.signer.Parties[id].Nonce
}

// GetPartyNoncePoint returns a party's nonce point
func (s *FrostAnimatedScene) GetPartyNoncePoint(id int) *secp256k1.PublicKey {
	if s.signer == nil || id < 0 || id >= len(s.signer.Parties) {
		return nil
	}
	return s.signer.Parties[id].NoncePoint
}

// GetPartyPartialSig returns a party's partial signature
func (s *FrostAnimatedScene) GetPartyPartialSig(id int) *big.Int {
	if s.signer == nil || id < 0 || id >= len(s.signer.Parties) {
		return nil
	}
	return s.signer.Parties[id].PartialSig
}
