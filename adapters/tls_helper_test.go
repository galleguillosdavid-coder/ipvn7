package adapters

import (
	"testing"
)

func TestGenerateTLSConfig(t *testing.T) {
	tlsConfig, err := GenerateTLSConfig()
	if err != nil {
		t.Fatalf("GenerateTLSConfig failed: %v", err)
	}

	if tlsConfig == nil {
		t.Fatal("GenerateTLSConfig returned nil tlsConfig")
	}

	if len(tlsConfig.Certificates) == 0 {
		t.Fatal("Expected at least one certificate in tlsConfig")
	}

	foundProto := false
	for _, proto := range tlsConfig.NextProtos {
		if proto == "ipv7-quic" {
			foundProto = true
			break
		}
	}

	if !foundProto {
		t.Errorf("Expected 'ipv7-quic' in NextProtos, got: %v", tlsConfig.NextProtos)
	}

	if !tlsConfig.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be true (IPv7 authenticates via Ed25519 identities)")
	}

	// Verify the certificate can be parsed
	cert := tlsConfig.Certificates[0]
	if len(cert.Certificate) == 0 {
		t.Fatal("Empty DER certificate slice")
	}
}
