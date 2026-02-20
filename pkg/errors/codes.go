package errors

type Code string

const (
	// Generic
	CodeInternal    Code = "INTERNAL"
	CodeUnavailable Code = "UNAVAILABLE"

	// Auth
	CodeUnauthenticated Code = "UNAUTHENTICATED"
	CodeForbidden       Code = "FORBIDDEN"

	// Validation / client
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeNotFound        Code = "NOT_FOUND"
	CodeConflict        Code = "CONFLICT"
	CodeTooManyRequests Code = "TOO_MANY_REQUESTS"
)

func (c Code) String() string {
	return string(c)
}
