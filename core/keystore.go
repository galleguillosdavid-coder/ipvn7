package core

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultKeyFileName defines the default identity key file name
const DefaultKeyFileName = "identity.key"

// DefaultKeyDir returns the standard application config directory (~/.ipv7)
func DefaultKeyDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ipv7"), nil
}

// LoadOrCreatePersistentIdentity loads an Ed25519 identity from disk or creates and saves a new one
func LoadOrCreatePersistentIdentity(customPath string) (*Ed25519Identity, ed25519.PrivateKey, error) {
	targetPath := customPath
	if targetPath == "" {
		dir, err := DefaultKeyDir()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, nil, fmt.Errorf("failed to create config dir %s: %w", dir, err)
		}
		targetPath = filepath.Join(dir, DefaultKeyFileName)
	}

	// 1. If key file exists, attempt to load seed
	if data, err := os.ReadFile(targetPath); err == nil {
		trimmed := strings.TrimSpace(string(data))
		seed, hexErr := hex.DecodeString(trimmed)
		if hexErr == nil && len(seed) == ed25519.SeedSize {
			privKey := ed25519.NewKeyFromSeed(seed)
			pubKey := privKey.Public().(ed25519.PublicKey)
			id, idErr := NewIdentityFromBytes(pubKey)
			if idErr == nil {
				return id, privKey, nil
			}
		}
	}

	// 2. Generate a new identity
	id, privKey, err := GenerateIdentity()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate new identity: %w", err)
	}

	// 3. Persist seed to disk
	seed := privKey.Seed()
	encoded := hex.EncodeToString(seed)
	if err := os.WriteFile(targetPath, []byte(encoded), 0600); err != nil {
		// Even if writing fails, return the valid generated identity
		return id, privKey, nil
	}

	return id, privKey, nil
}
