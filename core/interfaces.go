package core

// Identity represents the logical source/destination identifier (e.g. Ed25519 Public Key)
type Identity interface {
	// String representation of the identity (e.g. DID or base64)
	String() string
	
	// Bytes representation of the identity
	Bytes() []byte
	
	// Verify checks if the signature is valid for the given data
	Verify(data, signature []byte) bool
}

// Session represents a minimal state of communication
type Session interface {
	ID() string
	RemoteIdentity() Identity
}

// Adapter is the abstract interface for transport (UDP, QUIC, Relay)
type Adapter interface {
	// Start listening for incoming containers
	Start() error
	
	// Stop the adapter
	Stop() error
	
	// Send a container to a specific identity's endpoints
	Send(c *Container, endpoints []string) error
	
	// Receive channel for incoming containers
	Receive() <-chan *Container
}
