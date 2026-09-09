package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// KuzuExporter persists graph-relational telemetry events into Kùzu asynchronously
type KuzuExporter struct {
	kuzuExePath string
	dbPath      string
	logFile     *os.File
	mu          sync.Mutex
	inFlight    bool
}

// NewKuzuExporter initializes Kùzu persistence sink
func NewKuzuExporter(kuzuExePath, dbPath string) (*KuzuExporter, error) {
	if kuzuExePath == "" {
		kuzuExePath = filepath.Join("tools", "kuzu", "kuzu.exe")
	}
	if dbPath == "" {
		dbPath = filepath.Join(".kuzu_index", "ipv7_telemetry.db")
	}

	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)

	// Local JSONL journal for resilient append-only storage
	journalPath := filepath.Join(filepath.Dir(dbPath), "events_journal.jsonl")
	f, err := os.OpenFile(journalPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed opening telemetry journal: %w", err)
	}

	exp := &KuzuExporter{
		kuzuExePath: kuzuExePath,
		dbPath:      dbPath,
		logFile:     f,
	}

	// Initialize Kùzu Schema in background if kuzu.exe is available
	go exp.initSchema()

	return exp, nil
}

func (k *KuzuExporter) initSchema() {
	k.mu.Lock()
	defer k.mu.Unlock()

	schemaScript := `
CREATE NODE TABLE IF NOT EXISTS TelemetryNode(node_id STRING, PRIMARY KEY (node_id));
CREATE NODE TABLE IF NOT EXISTS TelemetryPeer(peer_id STRING, PRIMARY KEY (peer_id));
CREATE NODE TABLE IF NOT EXISTS TelemetrySession(session_id STRING, transport STRING, PRIMARY KEY (session_id));
CREATE NODE TABLE IF NOT EXISTS TelemetryEndpoint(addr STRING, PRIMARY KEY (addr));
CREATE NODE TABLE IF NOT EXISTS TelemetryEvent(event_id STRING, event_type STRING, timestamp INT64, rtt_ms DOUBLE, loss_pct DOUBLE, PRIMARY KEY (event_id));
CREATE REL TABLE IF NOT EXISTS HAS_PEER(FROM TelemetryNode TO TelemetryPeer);
CREATE REL TABLE IF NOT EXISTS HAS_SESSION(FROM TelemetryNode TO TelemetrySession);
CREATE REL TABLE IF NOT EXISTS USES_ENDPOINT(FROM TelemetrySession TO TelemetryEndpoint);
CREATE REL TABLE IF NOT EXISTS GENERATED(FROM TelemetryNode TO TelemetryEvent);
CREATE REL TABLE IF NOT EXISTS AFFECTED(FROM TelemetryEvent TO TelemetrySession);
`
	k.executeCypherScript(schemaScript)
}

// ConsumeBatch handles a batch of events asynchronously
func (k *KuzuExporter) ConsumeBatch(events []TelemetryEvent) error {
	if len(events) == 0 {
		return nil
	}

	// 1. Append immediately to resilient local JSONL journal (fast O(1) append)
	k.mu.Lock()
	for _, evt := range events {
		b, err := json.Marshal(evt)
		if err == nil {
			_, _ = k.logFile.Write(append(b, '\n'))
		}
	}
	k.mu.Unlock()

	// 2. Generate and trigger batch Cypher query in background
	go k.persistToKuzu(events)
	return nil
}

func (k *KuzuExporter) persistToKuzu(events []TelemetryEvent) {
	k.mu.Lock()
	if k.inFlight {
		k.mu.Unlock()
		return // Avoid stacking background sub-processes if previous is still busy
	}
	k.inFlight = true
	k.mu.Unlock()

	defer func() {
		k.mu.Lock()
		k.inFlight = false
		k.mu.Unlock()
	}()

	var sb strings.Builder
	for idx, evt := range events {
		evtID := fmt.Sprintf("%d_%d", evt.Timestamp, idx)
		sb.WriteString(fmt.Sprintf("MERGE (n:TelemetryNode {node_id: '%s'});\n", evt.NodeID))
		sb.WriteString(fmt.Sprintf("MERGE (e:TelemetryEvent {event_id: '%s', event_type: '%s', timestamp: %d, rtt_ms: %.3f, loss_pct: %.2f});\n",
			evtID, evt.Type, evt.Timestamp, evt.RTTMs, evt.LossPct))
		sb.WriteString(fmt.Sprintf("MERGE (n)-[:GENERATED]->(e);\n"))

		if evt.PeerID != "" {
			sb.WriteString(fmt.Sprintf("MERGE (p:TelemetryPeer {peer_id: '%s'});\n", evt.PeerID))
			sb.WriteString(fmt.Sprintf("MERGE (n)-[:HAS_PEER]->(p);\n"))
		}
		if evt.SessionID != "" {
			sb.WriteString(fmt.Sprintf("MERGE (s:TelemetrySession {session_id: '%s', transport: '%s'});\n", evt.SessionID, evt.Transport))
			sb.WriteString(fmt.Sprintf("MERGE (n)-[:HAS_SESSION]->(s);\n"))
			sb.WriteString(fmt.Sprintf("MERGE (e)-[:AFFECTED]->(s);\n"))

			if evt.NewEndpoint != "" {
				sb.WriteString(fmt.Sprintf("MERGE (ep:TelemetryEndpoint {addr: '%s'});\n", evt.NewEndpoint))
				sb.WriteString(fmt.Sprintf("MERGE (s)-[:USES_ENDPOINT]->(ep);\n"))
			}
		}
	}

	k.executeCypherScript(sb.String())
}

func (k *KuzuExporter) executeCypherScript(cypher string) {
	if _, err := os.Stat(k.kuzuExePath); err != nil {
		return // kuzu.exe not present; journal retains all data safely
	}

	cmd := exec.Command(k.kuzuExePath, k.dbPath, "-m", "json", "-s", "-b")
	cmd.Stdin = strings.NewReader(cypher + "\n")
	_ = cmd.Run() // Run silently, never panic or block
}

// Close closes the journal file
func (k *KuzuExporter) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.logFile != nil {
		return k.logFile.Close()
	}
	return nil
}
