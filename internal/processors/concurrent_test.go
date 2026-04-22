package processors

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/callmegreg/gh-security-config/internal/types"
)

// concurrencyTracker records how many concurrent calls happen at once.
type concurrencyTracker struct {
	mu        sync.Mutex
	current   int32
	maxSeen   int32
	holdFor   time.Duration
	results   map[string]types.ProcessingResult
	calledMu  sync.Mutex
	calledSet map[string]bool
}

func (c *concurrencyTracker) ProcessOrganization(org string) types.ProcessingResult {
	n := atomic.AddInt32(&c.current, 1)
	defer atomic.AddInt32(&c.current, -1)

	c.mu.Lock()
	if n > c.maxSeen {
		c.maxSeen = n
	}
	c.mu.Unlock()

	if c.holdFor > 0 {
		time.Sleep(c.holdFor)
	}

	c.calledMu.Lock()
	if c.calledSet == nil {
		c.calledSet = map[string]bool{}
	}
	c.calledSet[org] = true
	c.calledMu.Unlock()

	r, ok := c.results[org]
	if !ok {
		return types.ProcessingResult{Organization: org, Success: true}
	}
	r.Organization = org
	return r
}

func TestConcurrentProcessor_EmptyOrgs(t *testing.T) {
	p := NewConcurrentProcessor(nil, &fakeProcessor{}, 3)
	s, sk, e := p.Process()
	if s != 0 || sk != 0 || e != 0 {
		t.Errorf("expected all zero counts, got %d/%d/%d", s, sk, e)
	}
}

func TestConcurrentProcessor_CountsSuccessSkipAndError(t *testing.T) {
	fp := &fakeProcessor{results: map[string]types.ProcessingResult{
		"a": {Success: true},
		"b": {Success: true},
		"c": {Skipped: true},
		"d": {Error: errors.New("boom")},
	}}
	p := NewConcurrentProcessor([]string{"a", "b", "c", "d"}, fp, 2)
	s, sk, e := p.Process()
	if s != 2 || sk != 1 || e != 1 {
		t.Errorf("counts: success=%d skipped=%d errors=%d; want 2/1/1", s, sk, e)
	}
}

func TestConcurrentProcessor_RespectsConcurrencyLimit(t *testing.T) {
	ct := &concurrencyTracker{
		holdFor: 50 * time.Millisecond,
		results: map[string]types.ProcessingResult{},
	}
	orgs := []string{"a", "b", "c", "d", "e", "f"}
	p := NewConcurrentProcessor(orgs, ct, 2)
	p.Process()

	if ct.maxSeen > 2 {
		t.Errorf("max concurrency observed = %d, want <= 2", ct.maxSeen)
	}
	if ct.maxSeen < 1 {
		t.Errorf("expected at least 1 concurrent call, got %d", ct.maxSeen)
	}
}

func TestConcurrentProcessor_ConfigurationExistsTreatedAsSkip(t *testing.T) {
	fp := &fakeProcessor{results: map[string]types.ProcessingResult{
		"a": {Error: &types.ConfigurationExistsError{ConfigName: "cfg", OrgName: "a"}},
		"b": {Success: true},
	}}
	p := NewConcurrentProcessor([]string{"a", "b"}, fp, 2)
	s, sk, e := p.Process()
	if s != 1 || sk != 1 || e != 0 {
		t.Errorf("ConfigurationExistsError should be skip; got %d/%d/%d", s, sk, e)
	}
}

func TestConcurrentProcessor_DependabotUnavailableStopsProcessing(t *testing.T) {
	// Use a slow-holding processor so the dependabot error has time to signal stop
	// before all workers finish.
	orgs := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	fp := &fakeProcessor{results: map[string]types.ProcessingResult{
		"a": {Error: &types.DependabotUnavailableError{Feature: "alerts", OrgName: "a"}},
	}}
	p := NewConcurrentProcessor(orgs, fp, 1)
	s, sk, e := p.Process()

	total := s + sk + e
	if total != len(orgs) {
		t.Errorf("counts should sum to %d orgs, got %d (s=%d, sk=%d, e=%d)", len(orgs), total, s, sk, e)
	}
	if e < 1 {
		t.Errorf("expected at least one error, got %d", e)
	}
	if sk < 1 {
		t.Errorf("expected remaining orgs to be marked skipped, got %d", sk)
	}
}
