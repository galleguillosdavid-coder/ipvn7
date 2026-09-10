package tun

import (
	"io"
	"net"
	"sync"

	"ipv7/core"
)

// DefaultTunMTU standard MTU for the virtual interface to guarantee 0 fragmentation on WAN
const DefaultTunMTU = 1280

// TunDevice abstract interface for virtual network interfaces across OSes
type TunDevice interface {
	io.ReadWriteCloser
	Name() string
	MTU() int
}

// IPKey represents a fixed-size 16-byte key for zero-allocation route lookups
type IPKey [16]byte

// ToIPKey converts an IPv4 or IPv6 into a fixed 16-byte key without heap allocations
func ToIPKey(ip net.IP) IPKey {
	var key IPKey
	if ip4 := ip.To4(); ip4 != nil {
		copy(key[12:16], ip4)
	} else if len(ip) == 16 {
		copy(key[:], ip)
	}
	return key
}

// VirtualRouteTable maintains mapping between virtual IPs and sovereign DIDs with zero allocations
type VirtualRouteTable struct {
	mu      sync.RWMutex
	ipToDID map[IPKey]core.Identity
	didToIP map[string]net.IP
}

// NewVirtualRouteTable initializes an empty route table
func NewVirtualRouteTable() *VirtualRouteTable {
	return &VirtualRouteTable{
		ipToDID: make(map[IPKey]core.Identity),
		didToIP: make(map[string]net.IP),
	}
}

// Register registers an association between an IP and a DID
func (t *VirtualRouteTable) Register(ip net.IP, did core.Identity) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := ToIPKey(ip)
	t.ipToDID[key] = did
	t.didToIP[did.String()] = ip
}

// LookupDIDKey returns the DID associated with a fixed IPKey without allocations
func (t *VirtualRouteTable) LookupDIDKey(key IPKey) (core.Identity, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	did, found := t.ipToDID[key]
	return did, found
}

// LookupDID returns the DID associated with a virtual IP
func (t *VirtualRouteTable) LookupDID(ip net.IP) (core.Identity, bool) {
	return t.LookupDIDKey(ToIPKey(ip))
}

// LookupIP returns the virtual IP associated with a DID
func (t *VirtualRouteTable) LookupIP(did core.Identity) (net.IP, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ip, found := t.didToIP[did.String()]
	return ip, found
}
