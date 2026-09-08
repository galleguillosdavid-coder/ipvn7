package adapters

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// UPnPMapper handles lightweight UPnP IGD port mapping
type UPnPMapper struct {
	timeout time.Duration
}

// NewUPnPMapper creates a new mapper with default timeout
func NewUPnPMapper(timeout time.Duration) *UPnPMapper {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &UPnPMapper{timeout: timeout}
}

// DiscoverAndForward attempts to map the specified UDP port on the local IGD router
func (u *UPnPMapper) DiscoverAndForward(port int, desc string) error {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	// 1. Send SSDP M-SEARCH over UDP multicast
	multicastAddr, err := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	msearch := "M-SEARCH * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"ST: urn:schemas-upnp-org:service:WANIPConnection:1\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"MX: 2\r\n\r\n"

	_ = conn.SetDeadline(time.Now().Add(u.timeout))
	if _, err := conn.WriteTo([]byte(msearch), multicastAddr); err != nil {
		return err
	}

	buf := make([]byte, 2048)
	doneCh := make(chan string, 1)

	go func() {
		for {
			n, _, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}
			resp := string(buf[:n])
			for _, line := range strings.Split(resp, "\r\n") {
				if strings.HasPrefix(strings.ToUpper(line), "LOCATION:") {
					loc := strings.TrimSpace(line[len("LOCATION:"):])
					select {
					case doneCh <- loc:
					default:
					}
					return
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("upnp discovery timed out")
	case location := <-doneCh:
		// Discovered UPnP router device description URL
		return u.addPortMapping(ctx, location, port, desc)
	}
}

func (u *UPnPMapper) addPortMapping(ctx context.Context, location string, port int, desc string) error {
	// Simple validation of discovered URL
	if !strings.HasPrefix(location, "http://") {
		return fmt.Errorf("invalid upnp location: %s", location)
	}

	// Fetch XML description
	req, err := http.NewRequestWithContext(ctx, "GET", location, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Find controlURL for WANIPConnection
	bodyStr := string(body)
	controlURLTag := "<controlURL>"
	controlURLEndTag := "</controlURL>"
	idx := strings.Index(bodyStr, controlURLTag)
	if idx == -1 {
		return fmt.Errorf("controlURL not found in UPnP XML")
	}
	endIdx := strings.Index(bodyStr[idx:], controlURLEndTag)
	if endIdx == -1 {
		return fmt.Errorf("controlURL malformed")
	}

	controlPath := bodyStr[idx+len(controlURLTag) : idx+endIdx]
	var controlURL string
	if strings.HasPrefix(controlPath, "http://") {
		controlURL = controlPath
	} else {
		// Combine base URL from location
		parsed := strings.Split(location, "/")
		if len(parsed) >= 3 {
			baseURL := fmt.Sprintf("%s//%s", parsed[0], parsed[2])
			if !strings.HasPrefix(controlPath, "/") {
				controlPath = "/" + controlPath
			}
			controlURL = baseURL + controlPath
		}
	}

	if controlURL == "" {
		return fmt.Errorf("failed to build control URL")
	}

	// Determine local LAN IP
	lanConn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return err
	}
	defer lanConn.Close()
	localIP := lanConn.LocalAddr().(*net.UDPAddr).IP.String()

	// Send SOAP AddPortMapping
	soapBody := fmt.Sprintf(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>
<u:AddPortMapping xmlns:u="urn:schemas-upnp-org:service:WANIPConnection:1">
<NewRemoteHost></NewRemoteHost>
<NewExternalPort>%d</NewExternalPort>
<NewProtocol>UDP</NewProtocol>
<NewInternalPort>%d</NewInternalPort>
<NewInternalClient>%s</NewInternalClient>
<NewEnabled>1</NewEnabled>
<NewPortMappingDescription>%s</NewPortMappingDescription>
<NewLeaseDuration>3600</NewLeaseDuration>
</u:AddPortMapping>
</s:Body>
</s:Envelope>`, port, port, localIP, desc)

	soapReq, err := http.NewRequestWithContext(ctx, "POST", controlURL, bytes.NewBufferString(soapBody))
	if err != nil {
		return err
	}
	soapReq.Header.Set("Content-Type", "text/xml; charset=\"utf-8\"")
	soapReq.Header.Set("SOAPAction", "\"urn:schemas-upnp-org:service:WANIPConnection:1#AddPortMapping\"")

	soapResp, err := http.DefaultClient.Do(soapReq)
	if err != nil {
		return err
	}
	defer soapResp.Body.Close()

	if soapResp.StatusCode != http.StatusOK {
		return fmt.Errorf("upnp add port mapping returned status: %s", soapResp.Status)
	}

	return nil
}
