package auth

import "context"

// Decision represents the result of a policy evaluation.
type Decision int

const (
	// DecisionDeny means access is not allowed.
	DecisionDeny Decision = iota
	// DecisionAllow means access is allowed.
	DecisionAllow
)

// IsAllow reports whether decision is DecisionAllow.
func (d Decision) IsAllow() bool { return d == DecisionAllow }

// PolicyRequest is a normalized policy check input.
// Action/Resource should be stable strings, e.g. "tenant.members.invite", "orders.create".
type PolicyRequest struct {
	SubjectID string
	TenantID  string
	Scopes    []string

	Action   string
	Resource string
}

// PolicyChecker is implemented by a service adapter (HTTP/gRPC) that knows how to
// evaluate RBAC/ABAC rules (e.g. by calling Identity/Core).
type PolicyChecker interface {
	Check(ctx context.Context, req PolicyRequest) (Decision, error)
}
