package observability

import "testing"

func TestNewPrometheus_RegistryAndCollectors(t *testing.T) {
	t.Parallel()

	p := NewPrometheus()
	if p == nil || p.Registry == nil {
		t.Fatalf("expected non-nil prometheus registry")
	}

	mfs, err := p.Registry.Gather()
	if err != nil {
		t.Fatalf("gather err: %v", err)
	}
	if len(mfs) == 0 {
		t.Fatalf("expected some metrics to be registered")
	}

	// smoke check: expect at least one go_* or process_* metric family
	found := false
	for _, mf := range mfs {
		if mf.GetName() == "" {
			continue
		}
		if len(mf.GetName()) >= 3 && (mf.GetName()[:3] == "go_" || len(mf.GetName()) >= 8 && mf.GetName()[:8] == "process_") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected go_ or process_ metrics to be registered")
	}
}
