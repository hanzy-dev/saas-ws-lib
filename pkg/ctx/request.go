package ctx

import "context"

// WithRequestID returns a derived context carrying request_id.
//
// Defensive behavior: if ctx is nil, it is treated as context.Background().
// Empty requestID is ignored (ctx returned unchanged).
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		if ctx == nil {
			return context.Background()
		}
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, keyRequestID, requestID)
}

// RequestID returns the request_id stored in ctx, or empty string if not set.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, ok := ctx.Value(keyRequestID).(string)
	if !ok {
		return ""
	}
	return v
}

// WithTenantID returns a derived context carrying tenant_id.
//
// Defensive behavior: if ctx is nil, it is treated as context.Background().
// Empty tenantID is ignored (ctx returned unchanged).
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		if ctx == nil {
			return context.Background()
		}
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, keyTenantID, tenantID)
}

// TenantID returns the tenant_id stored in ctx, or empty string if not set.
func TenantID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, ok := ctx.Value(keyTenantID).(string)
	if !ok {
		return ""
	}
	return v
}

// WithSubjectID returns a derived context carrying subject_id (sub).
//
// Defensive behavior: if ctx is nil, it is treated as context.Background().
// Empty subjectID is ignored (ctx returned unchanged).
func WithSubjectID(ctx context.Context, subjectID string) context.Context {
	if subjectID == "" {
		if ctx == nil {
			return context.Background()
		}
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, keySubjectID, subjectID)
}

// SubjectID returns the subject_id stored in ctx, or empty string if not set.
func SubjectID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, ok := ctx.Value(keySubjectID).(string)
	if !ok {
		return ""
	}
	return v
}

// WithScopes returns a derived context carrying scopes ([]string).
//
// The slice is copied on write to prevent caller mutation.
// Defensive behavior: if ctx is nil, it is treated as context.Background().
// Empty scopes is ignored (ctx returned unchanged).
func WithScopes(ctx context.Context, scopes []string) context.Context {
	if len(scopes) == 0 {
		if ctx == nil {
			return context.Background()
		}
		return ctx
	}

	// make a copy to avoid accidental mutation by callers
	cp := make([]string, len(scopes))
	copy(cp, scopes)

	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, keyScopes, cp)
}

// Scopes returns the scopes stored in ctx.
//
// The returned slice is a copy to prevent caller mutation.
// If not set, returns nil.
func Scopes(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}
	v, ok := ctx.Value(keyScopes).([]string)
	if !ok || len(v) == 0 {
		return nil
	}

	// return a copy to prevent mutation
	cp := make([]string, len(v))
	copy(cp, v)
	return cp
}

// WithClaims stores arbitrary claims in context. Use sparingly.
// Defensive behavior: if ctx is nil, it is treated as context.Background().
func WithClaims[T any](ctx context.Context, claims T) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, keyClaims, claims)
}

// Claims loads claims of type T from ctx.
// Returns (zero, false) if not present or type mismatch.
func Claims[T any](ctx context.Context) (T, bool) {
	var zero T
	if ctx == nil {
		return zero, false
	}
	v, ok := ctx.Value(keyClaims).(T)
	if !ok {
		return zero, false
	}
	return v, true
}
