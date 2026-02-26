package ctx

// key is an unexported type to avoid collisions with other packages' context keys.
type key string

const (
	// Request-scoped identifiers
	keyRequestID key = "request_id"
	keyTenantID  key = "tenant_id"

	// Auth / identity
	keySubjectID key = "subject_id" // user/service id (sub)
	keyScopes    key = "scopes"     // []string

	// Optional: keep it lean; do not store large/untrusted payloads
	keyClaims key = "claims"
)
