package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var kuzuMu sync.Mutex

// FindKuzuBinary locates the Kùzu CLI executable based on OS and current working directory.
func FindKuzuBinary() string {
	var binName string
	if runtime.GOOS == "windows" {
		binName = "kuzu.exe"
	} else {
		binName = "kuzu"
	}

	candidates := []string{
		filepath.Join("tools", "kuzu", binName),
		filepath.Join("..", "tools", "kuzu", binName),
		filepath.Join("..", "..", "tools", "kuzu", binName),
	}

	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
			return c
		}
	}

	return binName // fallback to PATH
}

// FindKuzuDB locates the .kuzu_index/ipv7.db database path.
func FindKuzuDB() string {
	candidates := []string{
		filepath.Join(".kuzu_index", "ipv7.db"),
		filepath.Join("..", ".kuzu_index", "ipv7.db"),
		filepath.Join("..", "..", ".kuzu_index", "ipv7.db"),
	}

	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil {
			if fi.IsDir() || fi.Size() > 0 {
				abs, err := filepath.Abs(c)
				if err == nil {
					return abs
				}
				return c
			}
		}
	}

	return filepath.Join(".kuzu_index", "ipv7.db")
}

// ExecuteCypher runs a Cypher query on the Kùzu database using the Kùzu CLI and returns JSON output.
func ExecuteCypher(cypher string) ([]byte, error) {
	kuzuMu.Lock()
	defer kuzuMu.Unlock()

	kuzuBin := FindKuzuBinary()
	dbPath := FindKuzuDB()

	// Ensure db parent directory exists
	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)

	cmd := exec.Command(kuzuBin, dbPath, "-m", "json", "-s", "-b")
	cmd.Stdin = strings.NewReader(cypher + "\n")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("kuzu execution failed (%v): %s", err, stderr.String())
	}

	out := bytes.TrimSpace(stdout.Bytes())
	// In case multiple commands output multiple JSON blocks, take the last non-empty JSON block
	if bytes.HasPrefix(out, []byte("[")) {
		lastOpen := bytes.LastIndex(out, []byte("["))
		if lastOpen >= 0 {
			out = out[lastOpen:]
		}
	}

	return out, nil
}

// SyncPeersToKuzu writes the current node and peer network state into Kùzu.
func (s *Server) SyncPeersToKuzu() error {
	localID := s.node.Identity.String()
	localEp := fmt.Sprintf("127.0.0.1:%d", s.port)

	cypher := fmt.Sprintf("MERGE (p:Peer {id: '%s', endpoint: '%s', is_local: true});\n", localID, localEp)

	peers := s.node.SmallWorld.FindClosestPeers(s.node.Identity, 100)
	for _, p := range peers {
		pID := p.Identity.String()
		ep := "overlay"
		if len(p.Endpoints) > 0 {
			ep = p.Endpoints[0]
		}
		lat := p.Latency.Milliseconds()
		if lat <= 0 {
			lat = 1
		}

		cypher += fmt.Sprintf("MERGE (p:Peer {id: '%s', endpoint: '%s', is_local: false});\n", pID, ep)
		cypher += fmt.Sprintf("MATCH (a:Peer {id: '%s'}), (b:Peer {id: '%s'}) MERGE (a)-[:CONNECTED_TO {adapter: 'QUIC/UDP', latency_ms: %d, encrypted: true}]->(b);\n", localID, pID, lat)
	}

	_, err := ExecuteCypher(cypher)
	return err
}

// Handler for POST /api/kuzu/query
func (s *Server) handleKuzuQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	rawJSON, err := ExecuteCypher(req.Query)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(rawJSON)
}

// GraphNode for Cytoscape / D3 / HTML5 Canvas visualization
type GraphNode struct {
	ID        string                 `json:"id"`
	Label     string                 `json:"label"`
	Group     string                 `json:"group"` // "package", "file", "symbol", "peer", "local_node"
	Color     string                 `json:"color"`
	Radius    float64                `json:"radius"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// GraphLink for visual connection between nodes
type GraphLink struct {
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Label     string                 `json:"label"`
	Color     string                 `json:"color"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// GraphResponse for code or network visualizations
type GraphResponse struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

// Handler for GET /api/kuzu/network-graph
func (s *Server) handleKuzuNetworkGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Sync live peers into Kùzu
	_ = s.SyncPeersToKuzu()

	// 2. Query Kùzu for all peers and links
	query := `
MATCH (p:Peer)
OPTIONAL MATCH (p)-[r:CONNECTED_TO]->(m:Peer)
RETURN p.id, p.endpoint, p.is_local, r.adapter, r.latency_ms, r.encrypted, m.id;
`
	out, err := ExecuteCypher(query)
	if err != nil {
		// Fallback to in-memory mesh if Kuzu query fails
		s.handleMesh(w, r)
		return
	}

	type PeerQueryResult struct {
		PID       string  `json:"p.id"`
		PEndpoint string  `json:"p.endpoint"`
		PIsLocal  bool    `json:"p.is_local"`
		RAdapter  *string `json:"r.adapter"`
		RLatency  *int64  `json:"r.latency_ms"`
		REncrypt  *bool   `json:"r.encrypted"`
		MID       *string `json:"m.id"`
	}

	var rows []PeerQueryResult
	if err := json.Unmarshal(out, &rows); err != nil {
		s.handleMesh(w, r)
		return
	}

	nodesMap := make(map[string]GraphNode)
	var links []GraphLink

	for _, row := range rows {
		if row.PID != "" {
			if _, exists := nodesMap[row.PID]; !exists {
				group := "peer"
				color := "#8a2be2" // violet for remote peer
				radius := 14.0
				lbl := row.PID
				if len(lbl) > 10 {
					lbl = lbl[:10] + "..."
				}

				if row.PIsLocal {
					group = "local_node"
					color = "#00f2fe" // cyan for local
					radius = 20.0
					lbl = "Nodo Local (" + lbl + ")"
				}

				nodesMap[row.PID] = GraphNode{
					ID:     row.PID,
					Label:  lbl,
					Group:  group,
					Color:  color,
					Radius: radius,
					Metadata: map[string]interface{}{
						"endpoint": row.PEndpoint,
						"is_local": row.PIsLocal,
						"id":       row.PID,
					},
				}
			}
		}

		if row.MID != nil && *row.MID != "" {
			lat := int64(1)
			if row.RLatency != nil {
				lat = *row.RLatency
			}
			adapter := "QUIC/UDP"
			if row.RAdapter != nil {
				adapter = *row.RAdapter
			}
			enc := true
			if row.REncrypt != nil {
				enc = *row.REncrypt
			}

			links = append(links, GraphLink{
				Source: row.PID,
				Target: *row.MID,
				Label:  fmt.Sprintf("%s (%d ms)", adapter, lat),
				Color:  "rgba(0, 242, 254, 0.5)",
				Metadata: map[string]interface{}{
					"adapter":    adapter,
					"latency_ms": lat,
					"encrypted":  enc,
				},
			})
		}
	}

	var nodes []GraphNode
	for _, n := range nodesMap {
		nodes = append(nodes, n)
	}

	_ = json.NewEncoder(w).Encode(GraphResponse{
		Nodes: nodes,
		Links: links,
	})
}

// Handler for GET /api/kuzu/code-graph
func (s *Server) handleKuzuCodeGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := `
MATCH (p:Package)-[:CONTAINS]->(f:File)
OPTIONAL MATCH (f)-[:DEFINES]->(s:Symbol)
RETURN p.name, f.name, f.path, f.loc, s.name, s.kind
LIMIT 150;
`
	out, err := ExecuteCypher(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	type CodeQueryResult struct {
		PName *string `json:"p.name"`
		FName *string `json:"f.name"`
		FPath *string `json:"f.path"`
		FLoc  *int64  `json:"f.loc"`
		SName *string `json:"s.name"`
		SKind *string `json:"s.kind"`
	}

	var rows []CodeQueryResult
	if err := json.Unmarshal(out, &rows); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse Kuzu JSON: " + err.Error()})
		return
	}

	nodesMap := make(map[string]GraphNode)
	linksMap := make(map[string]GraphLink)

	pkgColors := map[string]string{
		"core":     "#00f2fe",
		"adapters": "#8a2be2",
		"dht":      "#10b981",
		"ui":       "#f59e0b",
		"main":     "#ec4899",
	}

	for _, row := range rows {
		if row.PName == nil || row.FName == nil {
			continue
		}
		pName := *row.PName
		fName := *row.FName
		fPath := *row.FPath
		fLoc := int64(0)
		if row.FLoc != nil {
			fLoc = *row.FLoc
		}

		// 1. Package node
		if _, exists := nodesMap["pkg:"+pName]; !exists {
			pColor, ok := pkgColors[pName]
			if !ok {
				pColor = "#6366f1"
			}
			nodesMap["pkg:"+pName] = GraphNode{
				ID:     "pkg:" + pName,
				Label:  "pkg: " + pName,
				Group:  "package",
				Color:  pColor,
				Radius: 24.0,
				Metadata: map[string]interface{}{
					"package": pName,
				},
			}
		}

		// 2. File node
		if _, exists := nodesMap["file:"+fPath]; !exists {
			nodesMap["file:"+fPath] = GraphNode{
				ID:     "file:" + fPath,
				Label:  fName,
				Group:  "file",
				Color:  "#38bdf8",
				Radius: 14.0,
				Metadata: map[string]interface{}{
					"package": pName,
					"file":    fName,
					"path":    fPath,
					"loc":     fLoc,
				},
			}
			// Link pkg -> file
			linkID := fmt.Sprintf("pkg:%s->file:%s", pName, fPath)
			linksMap[linkID] = GraphLink{
				Source: "pkg:" + pName,
				Target: "file:" + fPath,
				Label:  "contains",
				Color:  "rgba(255, 255, 255, 0.2)",
			}
		}

		// 3. Symbol node (if exists)
		if row.SName != nil && *row.SName != "" {
			sName := *row.SName
			sKind := "func"
			if row.SKind != nil {
				sKind = *row.SKind
			}
			sID := fmt.Sprintf("sym:%s::%s", fPath, sName)
			if _, exists := nodesMap[sID]; !exists {
				sColor := "#a78bfa"
				if sKind == "struct" {
					sColor = "#f43f5e"
				} else if sKind == "interface" {
					sColor = "#fbbf24"
				}

				nodesMap[sID] = GraphNode{
					ID:     sID,
					Label:  sName + " (" + sKind + ")",
					Group:  "symbol",
					Color:  sColor,
					Radius: 8.0,
					Metadata: map[string]interface{}{
						"name": sName,
						"kind": sKind,
						"file": fPath,
					},
				}

				linkID := fmt.Sprintf("file:%s->sym:%s", fPath, sID)
				linksMap[linkID] = GraphLink{
					Source: "file:" + fPath,
					Target: sID,
					Label:  "defines",
					Color:  "rgba(255, 255, 255, 0.1)",
				}
			}
		}
	}

	var nodes []GraphNode
	for _, n := range nodesMap {
		nodes = append(nodes, n)
	}

	var links []GraphLink
	for _, l := range linksMap {
		links = append(links, l)
	}

	_ = json.NewEncoder(w).Encode(GraphResponse{
		Nodes: nodes,
		Links: links,
	})
}
