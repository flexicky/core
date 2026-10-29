package grpcapp

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log/slog"

	"core/internal/service/token"
)

type keyPair struct {
	private ed25519.PrivateKey
	public  ed25519.PublicKey
}

// signingKeys returns the ed25519 key pair used for access tokens.
// A configured seed keeps tokens valid across restarts; without one a random
// pair is generated and every restart invalidates the issued tokens.
func signingKeys(encodedSeed string, log *slog.Logger) (keyPair, error) {
	if encodedSeed == "" {
		privateKey, publicKey, err := token.GenerateKeys()
		if err != nil {
			return keyPair{}, fmt.Errorf("generate signing keys: %w", err)
		}

		log.Warn("JWT_SEED is not set: a random signing key was generated, access tokens will not survive a restart")

		return keyPair{private: privateKey, public: publicKey}, nil
	}

	seed, err := base64.StdEncoding.DecodeString(encodedSeed)
	if err != nil {
		return keyPair{}, fmt.Errorf("decode jwt seed: %w", err)
	}

	privateKey, publicKey, err := token.KeysFromSeed(seed)
	if err != nil {
		return keyPair{}, fmt.Errorf("derive signing keys: %w", err)
	}

	return keyPair{private: privateKey, public: publicKey}, nil
}
