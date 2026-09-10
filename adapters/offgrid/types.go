package offgrid

import (
	"errors"
	"time"

	"ipv7/core"
)

var (
	ErrLinkClosed       = errors.New("physical link closed")
	ErrLinkTimeout      = errors.New("physical link read timeout")
	ErrBufferFull       = errors.New("physical link queue full")
	ErrPacketExceedsMTU = errors.New("packet exceeds physical link MTU")
)

// LinkMode defines whether a node is operating in global internet mode, pure off-grid, or hybrid
type LinkMode int

const (
	ModeOnlineInternet LinkMode = iota // Connected to WAN/Internet
	ModeOffGridPhysical                // Disconnected from Internet; running ad-hoc physical mesh
	ModeHybrid                         // Simultaneous online + local physical link
)

func (m LinkMode) String() string {
	switch m {
	case ModeOnlineInternet:
		return "ONLINE_INTERNET"
	case ModeOffGridPhysical:
		return "OFFGRID_PHYSICAL"
	case ModeHybrid:
		return "HYBRID"
	default:
		return "UNKNOWN"
	}
}

// PhysicalLink abstracts physical layer communication (Wi-Fi Direct, LoRa, ad-hoc radio)
type PhysicalLink interface {
	Name() string
	Send(targetAddr string, packet []byte) error
	Receive() (srcAddr string, packet []byte, err error)
	MTU() int
	Close() error
}

// PhysicalPeer represents an ad-hoc peer discovered directly via radio/layer-2 beacon
type PhysicalPeer struct {
	DID       core.Identity
	LocalAddr string
	RSSI      int // Signal strength in dBm (-30 to -120)
	LastSeen  time.Time
}
