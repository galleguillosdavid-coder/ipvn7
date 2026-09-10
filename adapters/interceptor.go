package adapters

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"ipv7/core"
)

// PacketAction defines the verdict of a smart packet filter.
type PacketAction int

const (
	ActionPass   PacketAction = iota // Continue normal packet processing
	ActionDrop                       // Silently drop the packet
	ActionModify                     // Replace packet with modified version
	ActionMirror                     // Replicate packet for asynchronous tap/telemetry
)

// HookPoint specifies where in the networking pipeline the filter executes.
type HookPoint string

const (
	HookIngress HookPoint = "INGRESS" // Arriving from physical interface
	HookEgress  HookPoint = "EGRESS"  // Before outbound transmission
	HookRelay   HookPoint = "RELAY"   // When acting as a transit hop
)

// FilterContext carries operational metadata across the filter pipeline.
type FilterContext struct {
	Hook       HookPoint
	PeerDID    string
	Timestamp  time.Time
	Timeout    time.Duration
	Attributes map[string]interface{}
}

// SmartFilter defines the contract for packet inspection, transformation, and sandboxed extensions.
type SmartFilter interface {
	Name() string
	Priority() int // Lower numbers execute earlier (e.g. 10 before 100)
	Process(ctx *FilterContext, c *core.Container) (PacketAction, *core.Container, error)
}

// PipelineStats tracks aggregated filter execution metrics.
type PipelineStats struct {
	TotalProcessed uint64
	TotalPassed    uint64
	TotalDropped   uint64
	TotalModified  uint64
	TotalMirrored  uint64
	TotalErrors    uint64
}

// FilterPipeline coordinates an ordered chain of SmartFilter plugins.
type FilterPipeline struct {
	mu      sync.RWMutex
	filters []SmartFilter
	stats   PipelineStats
	tapChan chan *core.Container
}

// NewFilterPipeline creates an empty pipeline with an optional mirror tap channel.
func NewFilterPipeline(tapBufferSize int) *FilterPipeline {
	var tap chan *core.Container
	if tapBufferSize > 0 {
		tap = make(chan *core.Container, tapBufferSize)
	}
	return &FilterPipeline{
		filters: make([]SmartFilter, 0),
		tapChan: tap,
	}
}

// Register adds a new filter and maintains priority ordering.
func (p *FilterPipeline) Register(filter SmartFilter) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.filters = append(p.filters, filter)
	sort.SliceStable(p.filters, func(i, j int) bool {
		return p.filters[i].Priority() < p.filters[j].Priority()
	})
}

// Unregister removes a filter by its identifier.
func (p *FilterPipeline) Unregister(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	updated := make([]SmartFilter, 0, len(p.filters))
	for _, f := range p.filters {
		if f.Name() != name {
			updated = append(updated, f)
		}
	}
	p.filters = updated
}

// Execute processes a packet sequentially through all registered filters with strict timeout.
func (p *FilterPipeline) Execute(hook HookPoint, peerDID string, c *core.Container) (PacketAction, *core.Container, error) {
	p.mu.RLock()
	filters := make([]SmartFilter, len(p.filters))
	copy(filters, p.filters)
	p.mu.RUnlock()

	p.mu.Lock()
	p.stats.TotalProcessed++
	p.mu.Unlock()

	ctx := &FilterContext{
		Hook:       hook,
		PeerDID:    peerDID,
		Timestamp:  time.Now(),
		Timeout:    10 * time.Millisecond, // Strict 10ms per-filter safety ceiling
		Attributes: make(map[string]interface{}),
	}

	currentContainer := c

	for _, filter := range filters {
		action, modified, err := executeWithTimeout(filter, ctx, currentContainer)
		if action == ActionDrop {
			p.mu.Lock()
			p.stats.TotalDropped++
			p.mu.Unlock()
			return ActionDrop, nil, nil
		}

		if err != nil {
			p.mu.Lock()
			p.stats.TotalErrors++
			p.mu.Unlock()
			// Fail-open for non-fatal filter errors to prevent packet starvation
			continue
		}

		case ActionModify:
			p.mu.Lock()
			p.stats.TotalModified++
			p.mu.Unlock()
			if modified != nil {
				currentContainer = modified
			}

		case ActionMirror:
			p.mu.Lock()
			p.stats.TotalMirrored++
			p.mu.Unlock()
			if p.tapChan != nil {
				select {
				case p.tapChan <- currentContainer:
				default: // Non-blocking tap: drop mirror if ring buffer is full
				}
			}

		case ActionPass:
			// Continue to next filter in chain
		}
	}

	p.mu.Lock()
	p.stats.TotalPassed++
	p.mu.Unlock()

	return ActionPass, currentContainer, nil
}

// executeWithTimeout wraps filter execution in a goroutine to enforce strict execution deadlines.
func executeWithTimeout(f SmartFilter, ctx *FilterContext, c *core.Container) (PacketAction, *core.Container, error) {
	type filterResult struct {
		action   PacketAction
		modified *core.Container
		err      error
	}

	resChan := make(chan filterResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resChan <- filterResult{action: ActionPass, err: fmt.Errorf("panic in filter %s: %v", f.Name(), r)}
			}
		}()
		act, mod, err := f.Process(ctx, c)
		resChan <- filterResult{action: act, modified: mod, err: err}
	}()

	timerCtx, cancel := context.WithTimeout(context.Background(), ctx.Timeout)
	defer cancel()

	select {
	case res := <-resChan:
		return res.action, res.modified, res.err
	case <-timerCtx.Done():
		return ActionPass, c, fmt.Errorf("filter %s exceeded timeout deadline (%v)", f.Name(), ctx.Timeout)
	}
}

// GetStats returns current pipeline execution statistics.
func (p *FilterPipeline) GetStats() PipelineStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// TapChannel exposes mirrored packets for external deep-packet inspection or telemetry.
func (p *FilterPipeline) TapChannel() <-chan *core.Container {
	return p.tapChan
}

// --- Built-in Smart Filters ---

// DLPFilter checks payload data for forbidden plain-text strings or signatures.
type DLPFilter struct {
	BlockedPatterns [][]byte
	priority        int
}

func NewDLPFilter(patterns []string, priority int) *DLPFilter {
	bPatterns := make([][]byte, len(patterns))
	for i, p := range patterns {
		bPatterns[i] = []byte(p)
	}
	return &DLPFilter{
		BlockedPatterns: bPatterns,
		priority:        priority,
	}
}

func (f *DLPFilter) Name() string { return "DLPFilter" }
func (f *DLPFilter) Priority() int { return f.priority }

func (f *DLPFilter) Process(ctx *FilterContext, c *core.Container) (PacketAction, *core.Container, error) {
	if c == nil || len(c.Payload) == 0 {
		return ActionPass, c, nil
	}

	for _, pattern := range f.BlockedPatterns {
		if bytes.Contains(c.Payload, pattern) {
			return ActionDrop, nil, errors.New("DLP signature detected: payload blocked")
		}
	}
	return ActionPass, c, nil
}

// MirrorFilter unconditionally replicates packets matching a predicate to the pipeline tap channel.
type MirrorFilter struct {
	TargetDID string
	priority  int
}

func NewMirrorFilter(targetDID string, priority int) *MirrorFilter {
	return &MirrorFilter{TargetDID: targetDID, priority: priority}
}

func (f *MirrorFilter) Name() string { return "MirrorFilter" }
func (f *MirrorFilter) Priority() int { return f.priority }

func (f *MirrorFilter) Process(ctx *FilterContext, c *core.Container) (PacketAction, *core.Container, error) {
	if ctx.PeerDID == f.TargetDID || f.TargetDID == "" {
		return ActionMirror, c, nil
	}
	return ActionPass, c, nil
}
