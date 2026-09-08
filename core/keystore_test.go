package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeystorePersistence(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "test_identity.key")

	// 1. First run: Generates and persists identity
	id1, priv1, err := LoadOrCreatePersistentIdentity(keyPath)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}
	if id1 == nil || priv1 == nil {
		t.Fatalf("expected non-nil identity and key")
	}

	// Verify file was written
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatalf("expected key file to exist at %s", keyPath)
	}

	// 2. Second run: Must reload identical identity
	id2, priv2, err := LoadOrCreatePersistentIdentity(keyPath)
	if err != nil {
		t.Fatalf("second load failed: %v", err)
	}

	if id1.String() != id2.String() {
		t.Fatalf("expected loaded identity %s to match persisted %s", id2.String(), id1.String())
	}
	if string(priv1) != string(priv2) {
		t.Fatalf("expected private keys to match exactly")
	}
}
