package auth

type Scope string

func (s Scope) String() string { return string(s) }

func Has(scopes []string, required Scope) bool {
	if len(scopes) == 0 {
		return false
	}
	want := required.String()
	for _, s := range scopes {
		if s == want {
			return true
		}
	}
	return false
}

func HasAll(scopes []string, required ...Scope) bool {
	for _, r := range required {
		if !Has(scopes, r) {
			return false
		}
	}
	return true
}

func HasAny(scopes []string, required ...Scope) bool {
	for _, r := range required {
		if Has(scopes, r) {
			return true
		}
	}
	return false
}
