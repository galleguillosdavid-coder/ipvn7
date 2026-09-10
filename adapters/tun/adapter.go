package tun

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"ipv7/core"
)

// TunAdapter connects a virtual network TUN interface with the IPv7 node overlay
type TunAdapter struct {
	device     TunDevice
	node       *core.Node
	routes     *VirtualRouteTable
	bufPool    sync.Pool
	running    atomic.Bool
	cancel     context.CancelFunc
	pktsIn     atomic.Uint64
	pktsOut    atomic.Uint64
	drops      atomic.Uint64
}

// NewTunAdapter creates a new TunAdapter instance
func NewTunAdapter(device TunDevice, node *core.Node, routes *VirtualRouteTable) *TunAdapter {
	if routes == nil {
		routes = NewVirtualRouteTable()
	}
	a := &TunAdapter{
		device: device,
		node:   node,
		routes: routes,
		bufPool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, device.MTU()+64)
				return &b
			},
		},
	}
	return a
}

// Start begins processing outbound IP packets from TUN and routing them into IPv7
func (a *TunAdapter) Start(ctx context.Context) error {
	if a.running.Swap(true) {
		return fmt.Errorf("tun adapter already running")
	}

	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	go a.readLoop(ctx)
	return nil
}

// Stop terminates the adapter
func (a *TunAdapter) Stop() error {
	if !a.running.Swap(false) {
		return nil
	}
	if a.cancel != nil {
		a.cancel()
	}
	return a.device.Close()
}

// Routes returns the virtual route table
func (a *TunAdapter) Routes() *VirtualRouteTable {
	return a.routes
}

// PacketsIn returns total packets read from the TUN interface
func (a *TunAdapter) PacketsIn() uint64 {
	return a.pktsIn.Load()
}

// PacketsOut returns total packets delivered to the TUN interface
func (a *TunAdapter) PacketsOut() uint64 {
	return a.pktsOut.Load()
}

// Drops returns total dropped packets
func (a *TunAdapter) Drops() uint64 {
	return a.drops.Load()
}

// DeliverInbound delivers an authentic IP packet received from an IPv7 peer to the OS TUN interface
func (a *TunAdapter) DeliverInbound(payload []byte) error {
	if !a.running.Load() {
		return fmt.Errorf("tun adapter not running")
	}
	_, err := a.device.Write(payload)
	if err == nil {
		a.pktsOut.Add(1)
	} else {
		a.drops.Add(1)
	}
	return err
}

// readLoop reads raw IP packets from the TUN device and routes them to the appropriate IPv7 peer
func (a *TunAdapter) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		bufPtr := a.bufPool.Get().(*[]byte)
		buf := *bufPtr

		n, err := a.device.Read(buf)
		if err != nil {
			a.bufPool.Put(bufPtr)
			if !a.running.Load() {
				return
			}
			continue
		}

		a.pktsIn.Add(1)
		packetData := buf[:n]

		// Parse destination IP directly into stack-allocated IPKey (0 allocs)
		dstKey, ok := extractDestinationIPKey(packetData)
		if !ok {
			a.drops.Add(1)
			a.bufPool.Put(bufPtr)
			continue
		}

		// Fast route lookup without heap allocations
		targetDID, found := a.routes.LookupDIDKey(dstKey)
		if !found {
			// Unknown virtual IP: drop packet or queue discovery
			a.drops.Add(1)
			a.bufPool.Put(bufPtr)
			continue
		}

		// Route over authenticated encrypted session to target DID
		go func(data []byte, did core.Identity, bPtr *[]byte) {
			defer a.bufPool.Put(bPtr)
			// Encapsulate packet payload and dispatch through IPv7 Node SendMessage
			_ = a.node.SendMessage(did, data)
		}(packetData, targetDID, bufPtr)
	}
}

// extractDestinationIPKey extracts destination IP as stack-allocated IPKey with 0 allocs
func extractDestinationIPKey(packet []byte) (IPKey, bool) {
	var key IPKey
	if len(packet) < 20 {
		return key, false
	}

	version := packet[0] >> 4
	if version == 4 {
		// IPv4 destination is at offset 16-20
		copy(key[12:16], packet[16:20])
		return key, true
	} else if version == 6 {
		if len(packet) < 40 {
			return key, false
		}
		// IPv6 destination is at offset 24-40
		copy(key[:], packet[24:40])
		return key, true
	}

	return key, false
}

// extractDestinationIP extracts IPv4 or IPv6 destination from raw IP packet (convenience)
func extractDestinationIP(packet []byte) net.IP {
	key, ok := extractDestinationIPKey(packet)
	if !ok {
		return nil
	}
	if packet[0]>>4 == 4 {
		return net.IPv4(key[12], key[13], key[14], key[15])
	}
	ip := make(net.IP, 16)
	copy(ip, key[:])
	return ip
}

// HandlePeerRoamed preserves the virtual IP while updating physical network endpoints
func (a *TunAdapter) HandlePeerRoamed(did core.Identity, newEndpoints []string) {
	// Virtual IP mapping remains strictly invariant
	// Underlying physical endpoint updated in core Node
	a.node.AddPeer(did, newEndpoints)
}

// ResolveAndRegister ensures a DID is mapped to an IPv6 ULA deterministically
func (a *TunAdapter) ResolveAndRegister(did core.Identity) net.IP {
	if ip, found := a.routes.LookupIP(did); found {
		return ip
	}
	ipv6 := DeriveIPv6FromDID(did)
	a.routes.Register(ipv6, did)
	return ipv6
}
