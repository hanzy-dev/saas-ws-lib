package httpx

import (
	"encoding/json"
	"testing"
)

func TestNewPage_EmptyItemsNotNil(t *testing.T) {
	p := NewPage[string](nil, "")

	if p.Items == nil {
		t.Fatalf("items must not be nil")
	}

	if len(p.Items) != 0 {
		t.Fatalf("expected empty slice")
	}

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := m["items"]; !ok {
		t.Fatalf("items field missing in json")
	}
}

func TestNewPage_WithItemsAndCursor(t *testing.T) {
	items := []int{1, 2}
	p := NewPage(items, "next123")

	if len(p.Items) != 2 {
		t.Fatalf("unexpected items length")
	}
	if p.NextCursor != "next123" {
		t.Fatalf("unexpected next cursor")
	}
}
