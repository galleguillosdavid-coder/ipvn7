package core

import (
	"encoding/json"
	"testing"
)

func TestMCPServer(t *testing.T) {
	id, priv, err := GenerateIdentity()
	if err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	node := NewNode(id, priv)
	mcp := NewMCPServer(node, nil)

	// 1. Test initialize
	initReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	raw, _ := json.Marshal(initReq)
	resp := mcp.HandleMessage(raw)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected successful initialize response, got: %v", resp)
	}

	// 2. Test tools/list
	listReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	raw, _ = json.Marshal(listReq)
	resp = mcp.HandleMessage(raw)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected tools list, got error: %v", resp)
	}

	// 3. Test tools/call: ipv7_get_info
	callParams, _ := json.Marshal(map[string]interface{}{
		"name":      "ipv7_get_info",
		"arguments": map[string]interface{}{},
	})
	callReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  callParams,
	}
	raw, _ = json.Marshal(callReq)
	resp = mcp.HandleMessage(raw)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected tools/call success, got: %v", resp)
	}

	// 4. Test tools/call: ipv7_query_kuzu with custom executor
	mcp.SetCypherExecutor(func(query string) (string, error) {
		return `[{"count": 5}]`, nil
	})
	kuzuParams, _ := json.Marshal(map[string]interface{}{
		"name": "ipv7_query_kuzu",
		"arguments": map[string]interface{}{
			"query": "MATCH (p:Package) RETURN count(p);",
		},
	})
	kuzuReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  kuzuParams,
	}
	raw, _ = json.Marshal(kuzuReq)
	resp = mcp.HandleMessage(raw)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected kuzu tool success, got: %v", resp)
	}

	// 5. Test tools/call: ipv7_toggle_vpn (start and stop)
	vpnParams, _ := json.Marshal(map[string]interface{}{
		"name": "ipv7_toggle_vpn",
		"arguments": map[string]interface{}{
			"action": "start",
			"port":   11080,
		},
	})
	vpnReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "tools/call",
		Params:  vpnParams,
	}
	raw, _ = json.Marshal(vpnReq)
	resp = mcp.HandleMessage(raw)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected vpn start success, got: %v", resp)
	}

	// Stop VPN
	vpnStopParams, _ := json.Marshal(map[string]interface{}{
		"name": "ipv7_toggle_vpn",
		"arguments": map[string]interface{}{
			"action": "stop",
		},
	})
	vpnStopReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "tools/call",
		Params:  vpnStopParams,
	}
	raw, _ = json.Marshal(vpnStopReq)
	resp = mcp.HandleMessage(raw)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected vpn stop success, got: %v", resp)
	}

	// 6. Test tools/call: ipv7_start_tunnel with attached tunnel service
	tunnelService := NewTunnelService(node)
	mcp = NewMCPServer(node, tunnelService)
	remoteID, _, _ := GenerateIdentity()

	tunnelParams, _ := json.Marshal(map[string]interface{}{
		"name": "ipv7_start_tunnel",
		"arguments": map[string]interface{}{
			"local_port":  15432,
			"peer_id":     remoteID.String(),
			"target_port": 5432,
		},
	})
	tunnelReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      7,
		Method:  "tools/call",
		Params:  tunnelParams,
	}
	raw, _ = json.Marshal(tunnelReq)
	resp = mcp.HandleMessage(raw)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected tunnel start success, got: %v", resp)
	}
	_ = tunnelService.CloseListener(15432)
}
