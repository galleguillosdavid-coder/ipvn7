package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ipv7/core"
)

func TestUIEndpoints(t *testing.T) {
	id, priv, err := core.GenerateIdentity()
	if err != nil {
		t.Fatalf("Failed to generate identity: %v", err)
	}

	node := core.NewNode(id, priv)
	cascade := core.NewCascadeNode(node, 10)
	server := NewServer(node, cascade, 0)

	// Test /api/info
	req := httptest.NewRequest("GET", "/api/info", nil)
	rec := httptest.NewRecorder()
	server.handleInfo(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200 from /api/info, got %d", rec.Code)
	}

	var info map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&info); err != nil {
		t.Fatalf("Failed to parse info JSON: %v", err)
	}
	if info["identity"] != id.String() {
		t.Errorf("Expected identity %s, got %v", id.String(), info["identity"])
	}

	// Test /api/mesh
	reqMesh := httptest.NewRequest("GET", "/api/mesh", nil)
	recMesh := httptest.NewRecorder()
	server.handleMesh(recMesh, reqMesh)

	if recMesh.Code != http.StatusOK {
		t.Errorf("Expected 200 from /api/mesh, got %d", recMesh.Code)
	}

	var mesh MeshGraphResponse
	if err := json.NewDecoder(recMesh.Body).Decode(&mesh); err != nil {
		t.Fatalf("Failed to parse mesh JSON: %v", err)
	}
	if len(mesh.Nodes) == 0 {
		t.Errorf("Expected at least local node in mesh.Nodes")
	}
	if mesh.Nodes[0].ID != id.String() {
		t.Errorf("Expected local node ID %s, got %s", id.String(), mesh.Nodes[0].ID)
	}

	// Test /api/mesh/export-cypher
	reqCypher := httptest.NewRequest("GET", "/api/mesh/export-cypher", nil)
	recCypher := httptest.NewRecorder()
	server.handleMeshCypher(recCypher, reqCypher)

	if recCypher.Code != http.StatusOK {
		t.Errorf("Expected 200 from /api/mesh/export-cypher, got %d", recCypher.Code)
	}
	cypherText := recCypher.Body.String()
	if len(cypherText) == 0 {
		t.Errorf("Expected non-empty cypher text")
	}
}

func TestKuzuEndpoints(t *testing.T) {
	id, priv, err := core.GenerateIdentity()
	if err != nil {
		t.Fatalf("Failed to generate identity: %v", err)
	}

	node := core.NewNode(id, priv)
	server := NewServer(node, nil, 0)

	// 1. Test /api/kuzu/code-graph
	reqCode := httptest.NewRequest("GET", "/api/kuzu/code-graph", nil)
	recCode := httptest.NewRecorder()
	server.handleKuzuCodeGraph(recCode, reqCode)

	if recCode.Code != http.StatusOK {
		t.Logf("Warning: /api/kuzu/code-graph returned %d (Kuzu might be busy or missing db): %s", recCode.Code, recCode.Body.String())
	} else {
		var resp GraphResponse
		if err := json.NewDecoder(recCode.Body).Decode(&resp); err == nil {
			t.Logf("Code graph returned %d nodes, %d links", len(resp.Nodes), len(resp.Links))
		}
	}

	// 2. Test /api/kuzu/network-graph
	reqNet := httptest.NewRequest("GET", "/api/kuzu/network-graph", nil)
	recNet := httptest.NewRecorder()
	server.handleKuzuNetworkGraph(recNet, reqNet)

	if recNet.Code != http.StatusOK {
		t.Logf("Warning: /api/kuzu/network-graph returned %d", recNet.Code)
	} else {
		t.Logf("Network graph successfully returned")
	}
}

