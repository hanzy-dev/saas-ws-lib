package ctx

import "context"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, KeyRequestID, requestID)
}

func RequestID(ctx context.Context) string {
	v, ok := ctx.Value(KeyRequestID).(string)
	if !ok {
		return ""
	}
	return v
}

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		return ctx
	}
	return context.WithValue(ctx, KeyTenantID, tenantID)
}

func TenantID(ctx context.Context) string {
	v, ok := ctx.Value(KeyTenantID).(string)
	if !ok {
		return ""
	}
	return v
}

func WithSubjectID(ctx context.Context, subjectID string) context.Context {
	if subjectID == "" {
		return ctx
	}
	return context.WithValue(ctx, KeySubjectID, subjectID)
}

func SubjectID(ctx context.Context) string {
	v, ok := ctx.Value(KeySubjectID).(string)
	if !ok {
		return ""
	}
	return v
}

func WithScopes(ctx context.Context, scopes []string) context.Context {
	if len(scopes) == 0 {
		return ctx
	}
	// make a copy to avoid accidental mutation by callers
	cp := make([]string, len(scopes))
	copy(cp, scopes)
	return context.WithValue(ctx, KeyScopes, cp)
}

func Scopes(ctx context.Context) []string {
	v, ok := ctx.Value(KeyScopes).([]string)
	if !ok || len(v) == 0 {
		return nil
	}
	// return a copy to prevent mutation
	cp := make([]string, len(v))
	copy(cp, v)
	return cp
}

func WithClaims[T any](ctx context.Context, claims T) context.Context {
	return context.WithValue(ctx, KeyClaims, claims)
}

func Claims[T any](ctx context.Context) (T, bool) {
	v, ok := ctx.Value(KeyClaims).(T)
	return v, ok
}
