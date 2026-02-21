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
	CodeInvalidArgument   Code = "INVALID_ARGUMENT"
	CodeNotFound          Code = "NOT_FOUND"
	CodeConflict          Code = "CONFLICT"
	CodeTooManyRequests   Code = "TOO_MANY_REQUESTS"
	CodeResourceExhausted Code = "RESOURCE_EXHAUSTED"

	// Optional (recommended for core/orders/payments)
	CodeDeadlineExceeded   Code = "DEADLINE_EXCEEDED"
	CodeAlreadyExists      Code = "ALREADY_EXISTS"
	CodeFailedPrecondition Code = "FAILED_PRECONDITION"
)

func (c Code) String() string {
	return string(c)
}
