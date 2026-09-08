package core

import (
	"sync"
	"time"
)

// RemoteInputEvent carries mouse and keyboard actions from the Web client
type RemoteInputEvent struct {
	Type   string  `json:"type"`   // "mousemove", "mousedown", "mouseup", "wheel", "keydown", "keyup"
	X      float64 `json:"x"`      // Normalized ratio (0.0 to 1.0)
	Y      float64 `json:"y"`      // Normalized ratio (0.0 to 1.0)
	Button int     `json:"button"` // 0: Left, 1: Middle, 2: Right
	DeltaY int     `json:"deltaY"` // Scroll delta
	Key    string  `json:"key"`    // Character or key name
	Code   string  `json:"code"`   // Key code
}

// RemoteDesktopService handles native lightweight screen capture and input injection
type RemoteDesktopService struct {
	mu           sync.Mutex
	quality      int
	width        int
	height       int
	scale        float64
	allowControl bool
}

// NewRemoteDesktopService initializes the remote desktop engine
func NewRemoteDesktopService() *RemoteDesktopService {
	rds := &RemoteDesktopService{
		quality:      70,
		scale:        0.8, // 80% scale for fast streaming
		allowControl: true,
	}
	rds.updateScreenMetrics()
	return rds
}

// GetResolution returns the current screen resolution
func (rds *RemoteDesktopService) GetResolution() (int, int) {
	rds.mu.Lock()
	defer rds.mu.Unlock()
	return rds.width, rds.height
}

// StartStreaming sends frames over a callback channel at target FPS
func (rds *RemoteDesktopService) StartStreaming(targetFPS int, frameCh chan<- []byte, stopCh <-chan struct{}) {
	if targetFPS <= 0 || targetFPS > 30 {
		targetFPS = 15 // Default 15 FPS is optimal for remote desktop
	}

	interval := time.Second / time.Duration(targetFPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			frame, err := rds.CaptureFrame()
			if err == nil && len(frame) > 0 {
				select {
				case frameCh <- frame:
				default:
					// Skip frame if consumer is congested
				}
			}
		}
	}
}
