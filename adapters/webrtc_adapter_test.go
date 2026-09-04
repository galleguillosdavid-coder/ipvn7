package adapters

import (
	"testing"
	"time"

	"ipv7/core"

	"github.com/pion/webrtc/v4"
)

func TestWebRTCAdapterCommunication(t *testing.T) {
	adapterA, err := NewWebRTCAdapter("")
	if err != nil {
		t.Fatalf("Failed to create adapter A: %v", err)
	}
	adapterB, err := NewWebRTCAdapter("")
	if err != nil {
		t.Fatalf("Failed to create adapter B: %v", err)
	}

	if err := adapterA.Start(); err != nil {
		t.Fatalf("Failed to start adapter A: %v", err)
	}
	defer adapterA.Stop()

	if err := adapterB.Start(); err != nil {
		t.Fatalf("Failed to start adapter B: %v", err)
	}
	defer adapterB.Stop()

	// Offerer creates data channel
	if err := adapterA.CreateDataChannel("ipv7-test"); err != nil {
		t.Fatalf("Failed to create data channel on A: %v", err)
	}

	// 1. A creates offer
	offer, err := adapterA.CreateOffer()
	if err != nil {
		t.Fatalf("Failed to create offer: %v", err)
	}

	// 2. B accepts offer and generates answer
	answer, err := adapterB.AcceptOffer(offer)
	if err != nil {
		t.Fatalf("Failed to accept offer on B: %v", err)
	}

	// 3. A accepts answer
	if err := adapterA.AcceptAnswer(answer); err != nil {
		t.Fatalf("Failed to accept answer on A: %v", err)
	}

	// Wait for DataChannel to open on A
	opened := make(chan struct{})
	adapterA.dataChan.OnOpen(func() {
		close(opened)
	})

	select {
	case <-opened:
	case <-time.After(5 * time.Second):
		if adapterA.dataChan.ReadyState() != webrtc.DataChannelStateOpen {
			t.Fatalf("Timed out waiting for WebRTC DataChannel to open (state: %s)", adapterA.dataChan.ReadyState())
		}
	}

	// 4. Create and sign Container
	idA, privA, _ := core.GenerateIdentity()
	payload := []byte("Hello via WebRTC DataChannel in IPv7!")
	c := &core.Container{
		SenderPubKey: idA.Bytes(),
		Payload:      payload,
	}
	if err := c.Sign(privA); err != nil {
		t.Fatalf("Failed to sign container: %v", err)
	}

	// 5. Send from A to B over WebRTC
	if err := adapterA.Send(c, nil); err != nil {
		t.Fatalf("Failed to send via WebRTC: %v", err)
	}

	// 6. Receive on B
	select {
	case received := <-adapterB.Receive():
		if string(received.Payload) != string(payload) {
			t.Fatalf("Payload mismatch: got %s, want %s", string(received.Payload), string(payload))
		}
		if !received.Verify() {
			t.Fatalf("Cryptographic verification failed on received WebRTC container")
		}
		t.Logf("Success! Received container over WebRTC DataChannel: %s", string(received.Payload))
	case <-time.After(3 * time.Second):
		t.Fatalf("Timed out waiting for container on WebRTC adapter B")
	}
}
