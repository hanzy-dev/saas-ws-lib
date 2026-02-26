package ctx

import (
	"context"
	"reflect"
	"testing"
)

func TestRequestID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
		in   string
		want string
	}{
		{"nil ctx + empty id", nil, "", ""},
		{"nil ctx + set id", nil, "r1", "r1"},
		{"bg ctx + empty id", context.Background(), "", ""},
		{"bg ctx + set id", context.Background(), "r2", "r2"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RequestID(WithRequestID(tt.ctx, tt.in))
			if got != tt.want {
				t.Fatalf("RequestID()=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestTenantID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
		in   string
		want string
	}{
		{"nil ctx + empty", nil, "", ""},
		{"nil ctx + set", nil, "t1", "t1"},
		{"bg ctx + empty", context.Background(), "", ""},
		{"bg ctx + set", context.Background(), "t2", "t2"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TenantID(WithTenantID(tt.ctx, tt.in))
			if got != tt.want {
				t.Fatalf("TenantID()=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestSubjectID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
		in   string
		want string
	}{
		{"nil ctx + empty", nil, "", ""},
		{"nil ctx + set", nil, "u1", "u1"},
		{"bg ctx + empty", context.Background(), "", ""},
		{"bg ctx + set", context.Background(), "u2", "u2"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SubjectID(WithSubjectID(tt.ctx, tt.in))
			if got != tt.want {
				t.Fatalf("SubjectID()=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestScopes_CopyOnWrite_And_CopyOnRead(t *testing.T) {
	t.Parallel()

	in := []string{"a", "b"}
	ctx1 := WithScopes(context.Background(), in)

	// mutate input slice after storing
	in[0] = "x"

	got1 := Scopes(ctx1)
	if !reflect.DeepEqual(got1, []string{"a", "b"}) {
		t.Fatalf("Scopes()=%v, want %v", got1, []string{"a", "b"})
	}

	// mutate returned slice and ensure ctx not affected
	got1[1] = "y"
	got2 := Scopes(ctx1)
	if !reflect.DeepEqual(got2, []string{"a", "b"}) {
		t.Fatalf("Scopes() after mutation=%v, want %v", got2, []string{"a", "b"})
	}

	// empty scopes should be ignored
	ctx2 := WithScopes(context.Background(), nil)
	if Scopes(ctx2) != nil {
		t.Fatalf("Scopes(empty) should be nil")
	}

	// ctx without scopes should return nil
	if Scopes(context.TODO()) != nil {
		t.Fatalf("Scopes(TODO) should be nil")
	}
}

func TestClaims_Generic(t *testing.T) {
	t.Parallel()

	type myClaims struct {
		Sub   string
		Admin bool
	}

	c := myClaims{Sub: "u1", Admin: true}
	ctx1 := WithClaims(context.Background(), c)

	got, ok := Claims[myClaims](ctx1)
	if !ok {
		t.Fatalf("Claims() ok=false, want true")
	}
	if got != c {
		t.Fatalf("Claims()=%v, want %v", got, c)
	}

	// type mismatch
	_, ok = Claims[string](ctx1)
	if ok {
		t.Fatalf("Claims(type mismatch) ok=true, want false")
	}

	// ctx without claims should return ok=false
	_, ok = Claims[myClaims](context.TODO())
	if ok {
		t.Fatalf("Claims(TODO) ok=true, want false")
	}
}
