package tests

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	mrand "math/rand/v2"
	"testing"
	"time"

	"ipv7/adapters/offgrid"
)

func TestConstrainedLink_StrictMTUEnforcement(t *testing.T) {
	medium := offgrid.NewRadioMedium()
	// Emulate severe LoRa link with MTU of 180 bytes
	link := offgrid.NewVirtualRadioLink("lora-180", "phy_lora_1", 180, medium)
	defer link.Close()

	smallPayload := make([]byte, 150)
	err := link.Send("phy_lora_2", smallPayload)
	if err != nil {
		t.Fatalf("Expected small packet <= MTU to succeed, got: %v", err)
	}

	largePayload := make([]byte, 512)
	err = link.Send("phy_lora_2", largePayload)
	if err != offgrid.ErrPacketExceedsMTU {
		t.Fatalf("Expected ErrPacketExceedsMTU for 512B packet on 180B MTU link, got: %v", err)
	}
	t.Log("Strict MTU enforcement verified: packets exceeding physical channel capacity rejected cleanly.")
}

func TestConstrainedLink_FragmentationAndReassemblyUnderChaos(t *testing.T) {
	const channelMTU = 180
	medium := offgrid.NewRadioMedium()

	txLink := offgrid.NewVirtualRadioLink("tx-radio", "phy_tx", channelMTU, medium)
	rxLink := offgrid.NewVirtualRadioLink("rx-radio", "phy_rx", channelMTU, medium)
	defer txLink.Close()
	defer rxLink.Close()

	// 1. Generate full-size 1280-byte IPv7 Sphinx Onion packet
	originalPacket := make([]byte, 1280)
	_, _ = rand.Read(originalPacket)
	originalHash := sha256.Sum256(originalPacket)

	// 2. Fragment packet for transmission over the 180-byte channel
	fragments, err := offgrid.FragmentPacket(originalPacket, channelMTU)
	if err != nil {
		t.Fatalf("FragmentPacket failed: %v", err)
	}

	expectedChunks := (1280 + (channelMTU - offgrid.FragHeaderSize) - 1) / (channelMTU - offgrid.FragHeaderSize)
	t.Logf("Paquete original de 1280B dividido en %d fragmentos (MTU=%dB, Overhead=%dB/frag)",
		len(fragments), channelMTU, offgrid.FragHeaderSize)

	if len(fragments) != expectedChunks {
		t.Fatalf("Expected %d fragments, got %d", expectedChunks, len(fragments))
	}

	// Verify all fragments strictly respect MTU
	for idx, frag := range fragments {
		if len(frag) > channelMTU {
			t.Fatalf("Fragment %d exceeds channel MTU: %d > %d", idx, len(frag), channelMTU)
		}
	}

	// 3. Inject Chaos: Shuffle fragment transmission order (Simulate multi-path jitter)
	shuffled := make([][]byte, len(fragments))
	copy(shuffled, fragments)
	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := mrand.IntN(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	t.Log("Transmitiendo fragmentos en orden estocástico desordenado...")
	for _, chunk := range shuffled {
		err := txLink.Send("phy_rx", chunk)
		if err != nil {
			t.Fatalf("txLink.Send error: %v", err)
		}
	}

	// 4. Receiver reassembly
	reassembler := offgrid.NewReassembler(2 * time.Second)
	var reconstructed []byte
	timeout := time.After(2 * time.Second)

	for reconstructed == nil {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for fragment reassembly under chaos")
		default:
			_, rxData, err := rxLink.Receive()
			if err != nil {
				t.Fatalf("rxLink.Receive error: %v", err)
			}

			full, complete, err := reassembler.ProcessChunk(rxData)
			if err != nil {
				t.Fatalf("ProcessChunk error: %v", err)
			}
			if complete {
				reconstructed = full
			}
		}
	}

	// 5. Cryptographic verification of bitwise integrity
	reconstructedHash := sha256.Sum256(reconstructed)
	if !bytes.Equal(originalPacket, reconstructed) {
		t.Fatalf("CORRUPCIÓN DE DATOS: Paquete reensamblado no coincide con el original de 1280B")
	}
	if reconstructedHash != originalHash {
		t.Fatalf("CORRUPCIÓN CRIPTOGRÁFICA: Hash SHA256 alterado tras reensamblado")
	}

	t.Logf("REENSAMBLADO EXITOSO: 1280 bytes íntegros (SHA256: %x...)", reconstructedHash[:6])
	t.Log("==========================================================================================")
	t.Log("RESULTADO FORMAL: HIPÓTESIS H-CONSTRAINED-MTU -> [DEMONSTRATED] EN LABORATORIO CONTROLADO")
	t.Log("Enlaces severamente restringidos (MTU=180B) transportan paquetes IPv7 de 1280B con 0 corrupción.")
	t.Log("==========================================================================================")
}
