package adapters

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/stun/v2"
)

// DefaultSTUNServer is a widely available public STUN server.
const DefaultSTUNServer = "stun.l.google.com:19302"

// DiscoverPublicEndpoint contacts a STUN server to discover the public reflexive IP and port.
func DiscoverPublicEndpoint(stunServer string) (string, error) {
	if stunServer == "" {
		stunServer = DefaultSTUNServer
	}

	// Dial UDP to the STUN server
	conn, err := net.Dial("udp", stunServer)
	if err != nil {
		return "", fmt.Errorf("failed to dial stun server %s: %w", stunServer, err)
	}
	defer conn.Close()

	client, err := stun.NewClient(conn)
	if err != nil {
		return "", fmt.Errorf("failed to create stun client: %w", err)
	}
	defer client.Close()

	var xorAddr stun.XORMappedAddress
	var xorErr error
	done := make(chan struct{})

	// Build STUN binding request
	req := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	err = client.Do(req, func(res stun.Event) {
		if res.Error != nil {
			xorErr = res.Error
			close(done)
			return
		}
		xorErr = xorAddr.GetFrom(res.Message)
		close(done)
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute stun request: %w", err)
	}

	select {
	case <-done:
		if xorErr != nil {
			return "", fmt.Errorf("failed to get XOR mapped address: %w", xorErr)
		}
		return fmt.Sprintf("%s:%d", xorAddr.IP.String(), xorAddr.Port), nil
	case <-time.After(3 * time.Second):
		return "", fmt.Errorf("stun request timed out")
	}
}

// GetLocalEndpoints returns non-loopback local IP addresses and the bound port.
func GetLocalEndpoints(port int) ([]string, error) {
	var endpoints []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.To4() != nil {
				endpoints = append(endpoints, fmt.Sprintf("%s:%d", ip.String(), port))
			}
		}
	}
	return endpoints, nil
}
