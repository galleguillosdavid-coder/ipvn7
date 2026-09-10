package offgrid

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"sync"
	"time"
)

var (
	FragMagic            = [2]byte{0x46, 0x37} // F7 (Fragmented IPv7 L2 frame)
	ErrCorruptedFragment = errors.New("corrupted fragment header")
	ErrReassemblyTimeout = errors.New("fragment reassembly timed out")
)

// FragHeaderSize: [FragMagic (2B)] [MsgID (2B)] [SeqNum (1B)] [Total (1B)]
const FragHeaderSize = 6

// FragmentPacket splits a large packet into chunks fitting within the physical MTU
func FragmentPacket(packet []byte, mtu int) ([][]byte, error) {
	if mtu <= FragHeaderSize {
		return nil, errors.New("MTU too small for fragmentation header")
	}
	payloadMax := mtu - FragHeaderSize

	totalFrags := (len(packet) + payloadMax - 1) / payloadMax
	if totalFrags > 255 {
		return nil, errors.New("packet too large to fragment within 255 chunks")
	}

	var msgID [2]byte
	_, _ = rand.Read(msgID[:])

	var chunks [][]byte
	for i := 0; i < totalFrags; i++ {
		start := i * payloadMax
		end := start + payloadMax
		if end > len(packet) {
			end = len(packet)
		}

		chunk := make([]byte, FragHeaderSize+(end-start))
		copy(chunk[:2], FragMagic[:])
		copy(chunk[2:4], msgID[:])
		chunk[4] = byte(i)
		chunk[5] = byte(totalFrags)
		copy(chunk[6:], packet[start:end])

		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

// Reassembler collects fragments and reconstructs the original packet
type Reassembler struct {
	sessions map[uint16]*reasmSession
	mu       sync.Mutex
	timeout  time.Duration
}

type reasmSession struct {
	total     int
	received  int
	parts     [][]byte
	createdAt time.Time
}

// NewReassembler creates a fragment collection manager with session expiration
func NewReassembler(timeout time.Duration) *Reassembler {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Reassembler{
		sessions: make(map[uint16]*reasmSession),
		timeout:  timeout,
	}
}

// ProcessChunk processes an incoming frame.
// Returns (originalPacket, isComplete, error).
// If the frame was not fragmented, it returns the frame as-is with isComplete = true.
func (r *Reassembler) ProcessChunk(chunk []byte) ([]byte, bool, error) {
	if len(chunk) < FragHeaderSize || !bytes.Equal(chunk[:2], FragMagic[:]) {
		// Non-fragmented frame, pass through
		return chunk, true, nil
	}

	msgID := binary.BigEndian.Uint16(chunk[2:4])
	seqNum := int(chunk[4])
	total := int(chunk[5])

	if seqNum >= total || total == 0 {
		return nil, false, ErrCorruptedFragment
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Evict timed out sessions
	now := time.Now()
	for id, sess := range r.sessions {
		if now.Sub(sess.createdAt) > r.timeout {
			delete(r.sessions, id)
		}
	}

	sess, exists := r.sessions[msgID]
	if !exists {
		sess = &reasmSession{
			total:     total,
			parts:     make([][]byte, total),
			createdAt: now,
		}
		r.sessions[msgID] = sess
	}

	if sess.parts[seqNum] == nil {
		payload := make([]byte, len(chunk)-FragHeaderSize)
		copy(payload, chunk[FragHeaderSize:])
		sess.parts[seqNum] = payload
		sess.received++
	}

	if sess.received == sess.total {
		delete(r.sessions, msgID)
		var fullPacket []byte
		for _, part := range sess.parts {
			fullPacket = append(fullPacket, part...)
		}
		return fullPacket, true, nil
	}

	return nil, false, nil
}
