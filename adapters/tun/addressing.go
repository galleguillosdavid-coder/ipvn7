package tun

import (
	"crypto/sha256"
	"net"

	"ipv7/core"
)

// DeriveIPv6FromDID derives a deterministic ULA IPv6 address (fd07::/64) from an Ed25519 DID
func DeriveIPv6FromDID(did core.Identity) net.IP {
	hash := sha256.Sum256([]byte(did.String()))

	ip := make(net.IP, 16)
	// ULA Prefix: fd07::/64
	ip[0] = 0xfd
	ip[1] = 0x07
	ip[2] = 0x00
	ip[3] = 0x00
	ip[4] = 0x00
	ip[5] = 0x00
	ip[6] = 0x00
	ip[7] = 0x00

	// 64 bits derived from SHA-256(DID)
	copy(ip[8:16], hash[0:8])

	return ip
}

// GenerateVirtualIPv4 generates a private IPv4 in 10.7.0.0/16 deterministically from an integer index
func GenerateVirtualIPv4(index uint16) net.IP {
	ip := make(net.IP, 4)
	ip[0] = 10
	ip[1] = 7
	ip[2] = byte((index >> 8) & 0xFF)
	ip[3] = byte(index & 0xFF)
	return ip
}
