package auth

import (
	"context"
	"testing"
)

type nopPolicy struct{}

func (nopPolicy) Check(ctx context.Context, req PolicyRequest) (Decision, error) {
	return DecisionAllow, nil
}

func TestDecision_IsAllow(t *testing.T) {
	t.Parallel()

	if !DecisionAllow.IsAllow() {
		t.Fatalf("DecisionAllow should be allow")
	}
	if DecisionDeny.IsAllow() {
		t.Fatalf("DecisionDeny should not be allow")
	}
}

func TestPolicyChecker_Interface(t *testing.T) {
	t.Parallel()

	var _ PolicyChecker = nopPolicy{}
}
