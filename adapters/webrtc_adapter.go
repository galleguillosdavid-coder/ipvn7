package adapters

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"ipv7/core"

	"github.com/pion/webrtc/v4"
)

// WebRTCAdapter implements core.Adapter over WebRTC RTCDataChannel
type WebRTCAdapter struct {
	api        *webrtc.API
	peerConn   *webrtc.PeerConnection
	dataChan   *webrtc.DataChannel
	receive    chan *core.Container
	stunServer string

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

// NewWebRTCAdapter creates a new WebRTC transport adapter
func NewWebRTCAdapter(stunServer string) (*WebRTCAdapter, error) {
	if stunServer == "" {
		stunServer = "stun:stun.l.google.com:19302"
	}

	return &WebRTCAdapter{
		receive:    make(chan *core.Container, 100),
		stunServer: stunServer,
		stopCh:     make(chan struct{}),
	}, nil
}

func (a *WebRTCAdapter) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return errors.New("webrtc adapter already running")
	}

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{a.stunServer},
			},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return fmt.Errorf("failed to create webrtc peer connection: %w", err)
	}

	a.peerConn = pc
	a.running = true
	a.stopCh = make(chan struct{})

	// Handle incoming remote data channels (when acting as answerer)
	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		a.mu.Lock()
		a.dataChan = dc
		a.mu.Unlock()
		a.bindDataChannel(dc)
	})

	return nil
}

func (a *WebRTCAdapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}
	a.running = false
	close(a.stopCh)

	if a.dataChan != nil {
		_ = a.dataChan.Close()
	}
	if a.peerConn != nil {
		return a.peerConn.Close()
	}
	return nil
}

// CreateDataChannel initializes an outgoing data channel (when acting as offerer)
func (a *WebRTCAdapter) CreateDataChannel(label string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.peerConn == nil {
		return errors.New("peer connection not started")
	}

	dc, err := a.peerConn.CreateDataChannel(label, nil)
	if err != nil {
		return err
	}

	a.dataChan = dc
	a.bindDataChannel(dc)
	return nil
}

func (a *WebRTCAdapter) bindDataChannel(dc *webrtc.DataChannel) {
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		c := &core.Container{}
		if err := c.Unmarshal(msg.Data); err != nil {
			return
		}
		if !c.Verify() {
			return
		}

		select {
		case a.receive <- c:
		default:
		}
	})
}

// CreateOffer generates a base64 encoded SDP offer string
func (a *WebRTCAdapter) CreateOffer() (string, error) {
	a.mu.Lock()
	pc := a.peerConn
	a.mu.Unlock()

	if pc == nil {
		return "", errors.New("peer connection not started")
	}

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return "", err
	}

	gatherComplete := webrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(offer); err != nil {
		return "", err
	}
	<-gatherComplete

	payload, err := json.Marshal(pc.LocalDescription())
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}

// AcceptOffer processes an SDP offer and returns a base64 encoded SDP answer string
func (a *WebRTCAdapter) AcceptOffer(b64Offer string) (string, error) {
	a.mu.Lock()
	pc := a.peerConn
	a.mu.Unlock()

	if pc == nil {
		return "", errors.New("peer connection not started")
	}

	raw, err := base64.StdEncoding.DecodeString(b64Offer)
	if err != nil {
		return "", fmt.Errorf("invalid base64 offer: %w", err)
	}

	var offer webrtc.SessionDescription
	if err := json.Unmarshal(raw, &offer); err != nil {
		return "", err
	}

	if err := pc.SetRemoteDescription(offer); err != nil {
		return "", err
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	gatherComplete := webrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(answer); err != nil {
		return "", err
	}
	<-gatherComplete

	payload, err := json.Marshal(pc.LocalDescription())
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}

// AcceptAnswer sets the remote description from an answered offer
func (a *WebRTCAdapter) AcceptAnswer(b64Answer string) error {
	a.mu.Lock()
	pc := a.peerConn
	a.mu.Unlock()

	if pc == nil {
		return errors.New("peer connection not started")
	}

	raw, err := base64.StdEncoding.DecodeString(b64Answer)
	if err != nil {
		return fmt.Errorf("invalid base64 answer: %w", err)
	}

	var answer webrtc.SessionDescription
	if err := json.Unmarshal(raw, &answer); err != nil {
		return err
	}

	return pc.SetRemoteDescription(answer)
}

func (a *WebRTCAdapter) Send(c *core.Container, endpoints []string) error {
	a.mu.Lock()
	dc := a.dataChan
	running := a.running
	a.mu.Unlock()

	if !running || dc == nil {
		return errors.New("webrtc data channel is not connected or open")
	}

	if dc.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("data channel is not open (state: %s)", dc.ReadyState().String())
	}

	data, err := c.Marshal()
	if err != nil {
		return err
	}

	return dc.Send(data)
}

func (a *WebRTCAdapter) Receive() <-chan *core.Container {
	return a.receive
}
