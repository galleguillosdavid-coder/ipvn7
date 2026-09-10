package dht

import (
	"bytes"
	"context"
	"testing"
	"time"

	"ipv7/core"
)

func TestXORDistanceMetric(t *testing.T) {
	idA, _, _ := core.GenerateIdentity()
	idB, _, _ := core.GenerateIdentity()
	idC, _, _ := core.GenerateIdentity()

	// 1. Identity: d(x, x) = 0
	dAA := XOR(idA.Bytes(), idA.Bytes())
	for _, b := range dAA {
		if b != 0 {
			t.Fatalf("Expected zero distance to self")
		}
	}

	// 2. Symmetry: d(x, y) = d(y, x)
	dAB := XOR(idA.Bytes(), idB.Bytes())
	dBA := XOR(idB.Bytes(), idA.Bytes())
	if !bytes.Equal(dAB[:], dBA[:]) {
		t.Fatalf("Symmetry failure in XOR distance")
	}

	// 3. Positivity: d(x, y) > 0 for x != y
	isZero := true
	for _, b := range dAB {
		if b != 0 {
			isZero = false
			break
		}
	}
	if isZero {
		t.Fatalf("Distance between distinct identities cannot be zero")
	}

	// 4. Triangle inequality: d(x, z) <= d(x, y) ^ d(y, z)
	dAC := XOR(idA.Bytes(), idC.Bytes())
	dBC := XOR(idB.Bytes(), idC.Bytes())
	dTri := XOR(dAB[:], dBC[:])
	if !bytes.Equal(dAC[:], dTri[:]) {
		t.Fatalf("Triangle equality in XOR metric")
	}
}

func TestPureKademliaTableKBuckets(t *testing.T) {
	localID, _, _ := core.GenerateIdentity()
	table := NewPureKademliaTable(localID)

	// Add 100 random contacts
	var addedIDs []core.Identity
	for i := 0; i < 100; i++ {
		cID, _, _ := core.GenerateIdentity()
		addedIDs = append(addedIDs, cID)
		table.AddContact(cID, []string{"192.168.1.1:7000"})
	}

	// Total contacts tracked will be bounded by k-bucket capacity (k=20 per prefix)
	total := table.TotalContacts()
	if total < 20 || total > 100 {
		t.Fatalf("Expected valid bounded contact count [20..100], got %d", total)
	}

	// Target query
	targetID, _, _ := core.GenerateIdentity()
	closest := table.FindClosest(targetID, 5)
	if len(closest) != 5 {
		t.Fatalf("Expected 5 closest contacts, got %d", len(closest))
	}

	// Verify monotonic order of distance
	prevDist := XOR(closest[0].ID.Bytes(), targetID.Bytes())
	for i := 1; i < len(closest); i++ {
		curDist := XOR(closest[i].ID.Bytes(), targetID.Bytes())
		if bytes.Compare(prevDist[:], curDist[:]) > 0 {
			t.Fatalf("Contacts not sorted monotonically by XOR distance")
		}
		prevDist = curDist
	}
}

func TestAntiSybilProofOfWork(t *testing.T) {
	payload := []byte("did:ipv7:alice_presence_record_payload")

	// Test difficulty 8 bits (1 zero byte)
	nonce := ComputePoW(payload, 8)
	if !VerifyPoW(payload, nonce, 8) {
		t.Fatalf("PoW verification failed for valid nonce")
	}

	// Tampered payload must fail
	if VerifyPoW([]byte("tampered_payload"), nonce, 8) {
		t.Fatalf("PoW verification succeeded unexpectedly for tampered payload")
	}
}

func TestDHTDiscoveryAdapterSovereign(t *testing.T) {
	// Node A
	idA, privA, _ := core.GenerateIdentity()
	nodeA := core.NewNode(idA, privA)
	dhtA := NewDHTService(idA, privA)
	adapterA := NewDHTDiscoveryAdapter(dhtA, nodeA)

	// Node B
	idB, privB, _ := core.GenerateIdentity()
	nodeB := core.NewNode(idB, privB)
	dhtB := NewDHTService(idB, privB)
	adapterB := NewDHTDiscoveryAdapter(dhtB, nodeB)

	// Connect DHT services via mock RPC router
	rpcRouter := func(targetEp string, req *Message) (*Message, error) {
		switch targetEp {
		case "ep-b":
			return dhtB.ProcessMessage(req), nil
		case "ep-a":
			return dhtA.ProcessMessage(req), nil
		default:
			return nil, nil
		}
	}
	dhtA.SetRemoteCaller(rpcRouter)
	dhtB.SetRemoteCaller(rpcRouter)

	dhtA.AddPeer(idB, []string{"ep-b"})
	dhtB.AddPeer(idA, []string{"ep-a"})

	nodeB.SetEndpoints([]string{"192.168.1.50:7002"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = adapterB.Start(ctx, time.Hour)
	defer adapterB.Stop()

	_ = adapterA.Start(ctx, time.Hour)
	defer adapterA.Stop()

	time.Sleep(50 * time.Millisecond)

	// Node A resolves Node B via DHT
	rec, err := dhtA.Resolve(idB.PublicKey)
	if err != nil {
		t.Fatalf("Failed to resolve Node B via sovereign DHT: %v", err)
	}

	if len(rec.Endpoints) != 1 || rec.Endpoints[0] != "192.168.1.50:7002" {
		t.Fatalf("Resolved endpoint mismatch: %v", rec.Endpoints)
	}

	t.Logf("Success! Resolved Node B without Firebase: %v", rec.Endpoints)
}

func BenchmarkXORDistance(b *testing.B) {
	idA, _, _ := core.GenerateIdentity()
	idB, _, _ := core.GenerateIdentity()
	a := idA.Bytes()
	c := idB.Bytes()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = XOR(a, c)
	}
}

func BenchmarkKademliaFindClosest(b *testing.B) {
	localID, _, _ := core.GenerateIdentity()
	table := NewPureKademliaTable(localID)

	// Populate 500 contacts
	for i := 0; i < 500; i++ {
		cID, _, _ := core.GenerateIdentity()
		table.AddContact(cID, []string{"127.0.0.1:7000"})
	}

	targetID, _, _ := core.GenerateIdentity()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = table.FindClosest(targetID, 10)
	}
}
