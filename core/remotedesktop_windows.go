//go:build windows

package core

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"syscall"
	"unsafe"
)

var (
	modUser32 = syscall.NewLazyDLL("user32.dll")
	modGdi32  = syscall.NewLazyDLL("gdi32.dll")

	procGetSystemMetrics       = modUser32.NewProc("GetSystemMetrics")
	procGetDC                  = modUser32.NewProc("GetDC")
	procReleaseDC              = modUser32.NewProc("ReleaseDC")
	procSetCursorPos           = modUser32.NewProc("SetCursorPos")
	procMouseEvent             = modUser32.NewProc("mouse_event")
	procKeybdEvent             = modUser32.NewProc("keybd_event")
	procVkKeyScanW             = modUser32.NewProc("VkKeyScanW")
	procCreateCompatibleDC     = modGdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = modGdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = modGdi32.NewProc("SelectObject")
	procBitBlt                 = modGdi32.NewProc("BitBlt")
	procGetDIBits              = modGdi32.NewProc("GetDIBits")
	procDeleteDC               = modGdi32.NewProc("DeleteDC")
	procDeleteObject           = modGdi32.NewProc("DeleteObject")
)

const (
	smCxScreen = 0
	smCyScreen = 1
	srccopy    = 0x00CC0020
	biRgb      = 0

	mouseeventfLeftDown   = 0x0002
	mouseeventfLeftUp     = 0x0004
	mouseeventfRightDown  = 0x0008
	mouseeventfRightUp    = 0x0010
	mouseeventfMiddleDown = 0x0020
	mouseeventfMiddleUp   = 0x0040
	mouseeventfWheel      = 0x0800
	keyeventfKeyup        = 0x0002
)

type bitmapInfoHeader struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

func (rds *RemoteDesktopService) updateScreenMetrics() {
	w, _, _ := procGetSystemMetrics.Call(smCxScreen)
	h, _, _ := procGetSystemMetrics.Call(smCyScreen)
	rds.mu.Lock()
	rds.width = int(w)
	rds.height = int(h)
	rds.mu.Unlock()
}

// CaptureFrame captures a single screenshot on Windows and encodes it as compressed JPEG bytes
func (rds *RemoteDesktopService) CaptureFrame() ([]byte, error) {
	rds.mu.Lock()
	width := rds.width
	height := rds.height
	quality := rds.quality
	rds.mu.Unlock()

	if width <= 0 || height <= 0 {
		rds.updateScreenMetrics()
		rds.mu.Lock()
		width = rds.width
		height = rds.height
		rds.mu.Unlock()
	}

	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return nil, fmt.Errorf("failed to get screen DC")
	}
	defer procReleaseDC.Call(0, hdc)

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	if memDC == 0 {
		return nil, fmt.Errorf("failed to create compatible DC")
	}
	defer procDeleteDC.Call(memDC)

	hBitmap, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(width), uintptr(height))
	if hBitmap == 0 {
		return nil, fmt.Errorf("failed to create compatible bitmap")
	}
	defer procDeleteObject.Call(hBitmap)

	oldBmp, _, _ := procSelectObject.Call(memDC, hBitmap)
	defer procSelectObject.Call(memDC, oldBmp)

	// Try capturing screen contents with SRCCOPY | CAPTUREBLT
	captureFlags := uintptr(0x00CC0020 | 0x40000000)
	ret, _, _ := procBitBlt.Call(memDC, 0, 0, uintptr(width), uintptr(height), hdc, 0, 0, captureFlags)
	if ret == 0 {
		// Retry without CAPTUREBLT
		ret, _, _ = procBitBlt.Call(memDC, 0, 0, uintptr(width), uintptr(height), hdc, 0, 0, 0x00CC0020)
	}

	if ret == 0 {
		// Fallback: render graceful standby frame (e.g. Session locked or non-interactive context)
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				idx := (y*width + x) * 4
				img.Pix[idx] = 15     // R
				img.Pix[idx+1] = 23   // G
				img.Pix[idx+2] = 42   // B
				img.Pix[idx+3] = 0xFF // A
			}
		}
		var buf bytes.Buffer
		_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		return buf.Bytes(), nil
	}

	var bi bitmapInfoHeader
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = int32(width)
	bi.BiHeight = -int32(height) // Top-down DIB
	bi.BiPlanes = 1
	bi.BiBitCount = 32
	bi.BiCompression = biRgb

	pixelBytes := make([]byte, width*height*4)
	ret, _, _ = procGetDIBits.Call(
		memDC,
		hBitmap,
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&pixelBytes[0])),
		uintptr(unsafe.Pointer(&bi)),
		0,
	)
	if ret == 0 {
		return nil, fmt.Errorf("getdibits failed")
	}

	// Windows DIB returns BGRA; convert to RGBA
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := 0; i < len(pixelBytes); i += 4 {
		img.Pix[i] = pixelBytes[i+2]   // R
		img.Pix[i+1] = pixelBytes[i+1] // G
		img.Pix[i+2] = pixelBytes[i]   // B
		img.Pix[i+3] = 0xFF            // A
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("jpeg encode failed: %w", err)
	}

	return buf.Bytes(), nil
}

// InjectInput executes a remote input action on Windows
func (rds *RemoteDesktopService) InjectInput(evt *RemoteInputEvent) error {
	if !rds.allowControl {
		return nil
	}

	rds.mu.Lock()
	w := rds.width
	h := rds.height
	rds.mu.Unlock()

	pixelX := int(evt.X * float64(w))
	pixelY := int(evt.Y * float64(h))

	switch evt.Type {
	case "mousemove":
		_, _, _ = procSetCursorPos.Call(uintptr(pixelX), uintptr(pixelY))

	case "mousedown":
		_, _, _ = procSetCursorPos.Call(uintptr(pixelX), uintptr(pixelY))
		var flag uintptr
		switch evt.Button {
		case 0:
			flag = mouseeventfLeftDown
		case 1:
			flag = mouseeventfMiddleDown
		case 2:
			flag = mouseeventfRightDown
		}
		if flag != 0 {
			_, _, _ = procMouseEvent.Call(flag, 0, 0, 0, 0)
		}

	case "mouseup":
		_, _, _ = procSetCursorPos.Call(uintptr(pixelX), uintptr(pixelY))
		var flag uintptr
		switch evt.Button {
		case 0:
			flag = mouseeventfLeftUp
		case 1:
			flag = mouseeventfMiddleUp
		case 2:
			flag = mouseeventfRightUp
		}
		if flag != 0 {
			_, _, _ = procMouseEvent.Call(flag, 0, 0, 0, 0)
		}

	case "wheel":
		delta := evt.DeltaY
		if delta > 0 {
			delta = -120
		} else {
			delta = 120
		}
		_, _, _ = procMouseEvent.Call(mouseeventfWheel, 0, 0, uintptr(delta), 0)

	case "keydown", "keyup":
		vk := getVirtualKey(evt.Key, evt.Code)
		if vk > 0 {
			var flags uintptr
			if evt.Type == "keyup" {
				flags = keyeventfKeyup
			}
			_, _, _ = procKeybdEvent.Call(uintptr(vk), 0, flags, 0)
		}
	}

	return nil
}

func getVirtualKey(key, code string) byte {
	switch key {
	case "Enter":
		return 0x0D
	case "Backspace":
		return 0x08
	case "Tab":
		return 0x09
	case "Escape":
		return 0x1B
	case "Space", " ":
		return 0x20
	case "ArrowLeft":
		return 0x25
	case "ArrowUp":
		return 0x26
	case "ArrowRight":
		return 0x27
	case "ArrowDown":
		return 0x28
	case "Delete":
		return 0x2E
	case "Shift":
		return 0x10
	case "Control":
		return 0x11
	case "Alt":
		return 0x12
	}

	if len(key) == 1 {
		char := rune(key[0])
		ret, _, _ := procVkKeyScanW.Call(uintptr(char))
		vk := byte(ret & 0xFF)
		if vk != 0xFF {
			return vk
		}
	}
	return 0
}
