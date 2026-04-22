package processors

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/types"
)

func init() {
	// Silence progress bar and info output during tests.
	pterm.DisableOutput()
}

// fakeProcessor returns a pre-programmed result per organization. It is safe for
// concurrent use so the same fake can back both sequential and concurrent tests.
type fakeProcessor struct {
	results map[string]types.ProcessingResult

	mu    sync.Mutex
	calls []string
}

func (f *fakeProcessor) ProcessOrganization(org string) types.ProcessingResult {
	f.mu.Lock()
	f.calls = append(f.calls, org)
	f.mu.Unlock()
	if r, ok := f.results[org]; ok {
		r.Organization = org
		return r
	}
	return types.ProcessingResult{Organization: org, Success: true}
}

func (f *fakeProcessor) callsSnapshot() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	copy(out, f.calls)
	return out
}

func TestSequentialProcessor_EmptyOrgs(t *testing.T) {
	p := NewSequentialProcessor(nil, &fakeProcessor{}, 0)
	s, sk, e := p.Process()
	if s != 0 || sk != 0 || e != 0 {
		t.Errorf("expected all zero counts, got success=%d skipped=%d errors=%d", s, sk, e)
	}
}

func TestSequentialProcessor_CountsSuccessSkipAndError(t *testing.T) {
	fp := &fakeProcessor{results: map[string]types.ProcessingResult{
		"a": {Success: true},
		"b": {Skipped: true},
		"c": {Error: errors.New("boom")},
	}}
	p := NewSequentialProcessor([]string{"a", "b", "c"}, fp, 0)
	s, sk, e := p.Process()
	if s != 1 || sk != 1 || e != 1 {
		t.Errorf("counts: success=%d skipped=%d errors=%d; want 1/1/1", s, sk, e)
	}
	if got := fp.callsSnapshot(); len(got) != 3 {
		t.Errorf("expected 3 processor calls, got %d", len(got))
	}
}

func TestSequentialProcessor_ConfigurationExistsTreatedAsSkip(t *testing.T) {
	fp := &fakeProcessor{results: map[string]types.ProcessingResult{
		"a": {Error: &types.ConfigurationExistsError{ConfigName: "cfg", OrgName: "a"}},
		"b": {Success: true},
	}}
	p := NewSequentialProcessor([]string{"a", "b"}, fp, 0)
	s, sk, e := p.Process()
	if s != 1 || sk != 1 || e != 0 {
		t.Errorf("ConfigurationExistsError should be counted as skip; got success=%d skipped=%d errors=%d", s, sk, e)
	}
}

func TestSequentialProcessor_DependabotUnavailableStopsProcessing(t *testing.T) {
	fp := &fakeProcessor{results: map[string]types.ProcessingResult{
		"a": {Success: true},
		"b": {Error: &types.DependabotUnavailableError{Feature: "alerts", OrgName: "b"}},
		// c and d should not be called but are recorded as skipped.
	}}
	p := NewSequentialProcessor([]string{"a", "b", "c", "d"}, fp, 0)
	s, sk, e := p.Process()
	if s != 1 {
		t.Errorf("success: got %d, want 1", s)
	}
	if e != 1 {
		t.Errorf("errors: got %d, want 1", e)
	}
	// c and d should be marked as skipped.
	if sk != 2 {
		t.Errorf("skipped: got %d, want 2 (remaining orgs)", sk)
	}
	// c and d must not be processed.
	for _, called := range fp.callsSnapshot() {
		if called == "c" || called == "d" {
			t.Errorf("processor should not have been called for %q after dependabot error", called)
		}
	}
}

func TestSequentialProcessor_DelayBetweenOrgs(t *testing.T) {
	fp := &fakeProcessor{}
	// 1-second delay between 2 orgs -> expect at least ~1s elapsed.
	p := NewSequentialProcessor([]string{"a", "b"}, fp, 1)
	start := time.Now()
	p.Process()
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Errorf("expected delay between orgs to take ~1s, got %s", elapsed)
	}
}
