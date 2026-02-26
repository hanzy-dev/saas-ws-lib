package auth

import "testing"

func TestHas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scopes   []string
		required Scope
		want     bool
	}{
		{"nil scopes", nil, "a", false},
		{"empty required", []string{"a"}, "", false},
		{"missing", []string{"a"}, "b", false},
		{"match exact", []string{"a", "b"}, "b", true},
		{"trim match", []string{"  a  "}, "a", true},
		{"required trimmed", []string{"a"}, "  a  ", true},
		{"ignore empty scopes entries", []string{"", "a"}, "a", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Has(tt.scopes, tt.required); got != tt.want {
				t.Fatalf("Has()=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scopes   []string
		required []Scope
		want     bool
	}{
		{"no required => true", []string{"a"}, nil, true},
		{"only empty required => true", []string{"a"}, []Scope{""}, true},
		{"missing one => false", []string{"a"}, []Scope{"a", "b"}, false},
		{"all present => true", []string{"a", "b"}, []Scope{"a", "b"}, true},
		{"duplicates in required => true", []string{"a"}, []Scope{"a", "a"}, true},
		{"trim + ignore empty => true", []string{"a"}, []Scope{"  a  ", ""}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasAll(tt.scopes, tt.required...); got != tt.want {
				t.Fatalf("HasAll()=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scopes   []string
		required []Scope
		want     bool
	}{
		{"no required => false", []string{"a"}, nil, false},
		{"only empty required => false", []string{"a"}, []Scope{""}, false},
		{"none match => false", []string{"a"}, []Scope{"b", "c"}, false},
		{"one match => true", []string{"a", "b"}, []Scope{"c", "b"}, true},
		{"trim match => true", []string{"  a  "}, []Scope{"a"}, true},
		{"duplicates required => true", []string{"a"}, []Scope{"a", "a"}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasAny(tt.scopes, tt.required...); got != tt.want {
				t.Fatalf("HasAny()=%v, want %v", got, tt.want)
			}
		})
	}
}
