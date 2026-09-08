package core

import (
	"bytes"
	"image/jpeg"
	"testing"
)

func TestRemoteDesktopCapture(t *testing.T) {
	rds := NewRemoteDesktopService()

	w, h := rds.GetResolution()
	if w <= 0 || h <= 0 {
		t.Fatalf("Invalid resolution: %dx%d", w, h)
	}
	t.Logf("Detected display resolution: %dx%d", w, h)

	frame, err := rds.CaptureFrame()
	if err != nil {
		t.Fatalf("Failed to capture frame: %v", err)
	}

	if len(frame) == 0 {
		t.Fatalf("Captured empty frame")
	}

	// Verify that frame is valid JPEG
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(frame))
	if err != nil {
		t.Fatalf("Frame is not a valid JPEG: %v", err)
	}

	t.Logf("Successfully captured JPEG frame: %dx%d (%d KB)", cfg.Width, cfg.Height, len(frame)/1024)
}
