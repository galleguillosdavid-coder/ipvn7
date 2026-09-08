//go:build !windows

package core

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
)

func (rds *RemoteDesktopService) updateScreenMetrics() {
	rds.mu.Lock()
	rds.width = 1920
	rds.height = 1080
	rds.mu.Unlock()
}

// CaptureFrame provides a graceful virtual framebuffer for Linux / WSL2 headless environments
func (rds *RemoteDesktopService) CaptureFrame() ([]byte, error) {
	rds.mu.Lock()
	width := rds.width
	height := rds.height
	quality := rds.quality
	rds.mu.Unlock()

	if width <= 0 || height <= 0 {
		width = 1280
		height = 720
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with dark background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{R: 15, G: 23, B: 42, A: 255}}, image.Point{}, draw.Src)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// InjectInput is a graceful no-op on non-Windows headless environments
func (rds *RemoteDesktopService) InjectInput(evt *RemoteInputEvent) error {
	return nil
}
