package ctx

// key is an unexported type to avoid collisions with other packages' context keys.
type key string

const (
	// Request-scoped identifiers
	KeyRequestID key = "request_id"
	KeyTenantID  key = "tenant_id"

	// Auth / identity
	KeySubjectID key = "subject_id" // user/service id (sub)
	KeyScopes    key = "scopes"     // []string

	// Optional: store raw token claims if you want (keep it lean; don't overuse)
	KeyClaims key = "claims"
)
