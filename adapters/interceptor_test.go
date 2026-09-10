package adapters

import (
	"testing"
	"time"

	"ipv7/core"
)

func TestPipeline_Pass(t *testing.T) {
	pipeline := NewFilterPipeline(10)
	c := &core.Container{
		Payload: []byte("Hello clean world!"),
	}

	action, out, err := pipeline.Execute(HookIngress, "did:ipv7:alice", c)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if action != ActionPass || out == nil {
		t.Fatalf("Expected ActionPass and non-nil container, got action=%v", action)
	}

	stats := pipeline.GetStats()
	if stats.TotalProcessed != 1 || stats.TotalPassed != 1 {
		t.Fatalf("Stats mismatch: %+v", stats)
	}
}

func TestPipeline_DLP_Drop(t *testing.T) {
	pipeline := NewFilterPipeline(10)
	dlp := NewDLPFilter([]string{"CONFIDENTIAL_LEAK", "PRIVATE_KEY_BYTES"}, 10)
	pipeline.Register(dlp)

	// Clean container -> PASS
	clean := &core.Container{
		Payload: []byte("Safe public message"),
	}
	action, _, err := pipeline.Execute(HookEgress, "did:ipv7:bob", clean)
	if err != nil || action != ActionPass {
		t.Fatalf("Clean container should pass, got action=%v, err=%v", action, err)
	}

	// Dirty container -> DROP
	dirty := &core.Container{
		Payload: []byte("Exporting secret CONFIDENTIAL_LEAK to peer"),
	}
	action, out, _ := pipeline.Execute(HookEgress, "did:ipv7:bob", dirty)
	if action != ActionDrop || out != nil {
		t.Fatalf("Dirty container should be dropped by DLPFilter, got action=%v", action)
	}

	stats := pipeline.GetStats()
	if stats.TotalDropped != 1 {
		t.Fatalf("Expected 1 dropped packet in stats, got %d", stats.TotalDropped)
	}
}

func TestPipeline_MirrorTap(t *testing.T) {
	pipeline := NewFilterPipeline(5)
	mirror := NewMirrorFilter("did:ipv7:target_peer", 20)
	pipeline.Register(mirror)

	c := &core.Container{
		Payload: []byte("Mirrored packet payload"),
	}

	action, _, err := pipeline.Execute(HookIngress, "did:ipv7:target_peer", c)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if action != ActionPass {
		t.Fatalf("Packet should pass along to receiver even when mirrored")
	}

	select {
	case tapPkt := <-pipeline.TapChannel():
		if string(tapPkt.Payload) != "Mirrored packet payload" {
			t.Fatalf("Tap payload mismatch: %s", string(tapPkt.Payload))
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Timed out waiting for packet on mirror tap channel")
	}
}

type hangingFilter struct{}

func (h *hangingFilter) Name() string { return "hangingFilter" }
func (h *hangingFilter) Priority() int { return 1 }
func (h *hangingFilter) Process(ctx *FilterContext, c *core.Container) (PacketAction, *core.Container, error) {
	// Hang deliberately longer than context timeout
	time.Sleep(50 * time.Millisecond)
	return ActionPass, c, nil
}

func TestPipeline_TimeoutProtection(t *testing.T) {
	pipeline := NewFilterPipeline(0)
	pipeline.Register(&hangingFilter{})

	c := &core.Container{
		Payload: []byte("Quick payload"),
	}

	start := time.Now()
	action, _, _ := pipeline.Execute(HookIngress, "did:ipv7:charlie", c)
	elapsed := time.Since(start)

	// Context timeout is 10ms; the hanging filter takes 50ms.
	// The pipeline must abort the hanging filter and fail-open in ~15ms, not 50ms+.
	if elapsed > 35*time.Millisecond {
		t.Fatalf("Pipeline failed to enforce strict execution deadline, elapsed: %v", elapsed)
	}

	if action != ActionPass {
		t.Fatalf("Expected fail-open ActionPass on timeout, got action=%v", action)
	}

	stats := pipeline.GetStats()
	if stats.TotalErrors != 1 {
		t.Fatalf("Expected 1 recorded error for timeout, got %d", stats.TotalErrors)
	}
}
