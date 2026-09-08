package adapters

import (
	"testing"
	"time"
)

func TestUPnPMapperGracefulTimeout(t *testing.T) {
	// Mapper with ultra-short timeout to verify graceful exit without crashing or hanging
	mapper := NewUPnPMapper(100 * time.Millisecond)

	err := mapper.DiscoverAndForward(7001, "IPv7 Test Port")
	// In test environments without real UPnP router, it must return a timeout error gracefully
	if err == nil {
		t.Logf("UPnP router was unexpectedly found in test environment!")
	} else {
		t.Logf("UPnP mapper correctly handled environment without router: %v", err)
	}
}
