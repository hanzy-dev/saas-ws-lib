package auth

import "strings"

// Scope represents an authorization scope (permission string) attached to an identity.
type Scope string

func (s Scope) String() string { return string(s) }

// Has reports whether scopes contains the required scope.
// Matching is exact after TrimSpace normalization.
func Has(scopes []string, required Scope) bool {
	req := strings.TrimSpace(required.String())
	if req == "" || len(scopes) == 0 {
		return false
	}
	for _, s := range scopes {
		if strings.TrimSpace(s) == req {
			return true
		}
	}
	return false
}

// HasAll reports whether scopes contains all required scopes.
// Empty required values are ignored.
// If no non-empty required scopes are provided, returns true.
func HasAll(scopes []string, required ...Scope) bool {
	reqs := normalizeRequired(required...)
	if len(reqs) == 0 {
		return true
	}
	set := scopeSet(scopes)
	for r := range reqs {
		if !set[r] {
			return false
		}
	}
	return true
}

// HasAny reports whether scopes contains any of the required scopes.
// Empty required values are ignored.
// If no non-empty required scopes are provided, returns false.
func HasAny(scopes []string, required ...Scope) bool {
	reqs := normalizeRequired(required...)
	if len(reqs) == 0 {
		return false
	}
	set := scopeSet(scopes)
	for r := range reqs {
		if set[r] {
			return true
		}
	}
	return false
}

func scopeSet(scopes []string) map[string]bool {
	if len(scopes) == 0 {
		return map[string]bool{}
	}
	set := make(map[string]bool, len(scopes))
	for _, s := range scopes {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		set[s] = true
	}
	return set
}

func normalizeRequired(required ...Scope) map[string]bool {
	if len(required) == 0 {
		return nil
	}
	out := make(map[string]bool, len(required))
	for _, r := range required {
		s := strings.TrimSpace(r.String())
		if s == "" {
			continue
		}
		out[s] = true
	}
	return out
}
