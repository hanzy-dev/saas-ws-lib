package validate

import (
	"testing"

	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

type createUserReq struct {
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=18"`
}

func TestStruct_Valid(t *testing.T) {
	req := createUserReq{
		Email: "user@example.com",
		Age:   20,
	}

	err := Struct(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStruct_Invalid(t *testing.T) {
	req := createUserReq{
		Email: "invalid",
		Age:   15,
	}

	err := Struct(req)
	if err == nil {
		t.Fatalf("expected validation error")
	}

	if err.Code != wserr.CodeInvalidArgument {
		t.Fatalf("code mismatch: got=%s want=%s", err.Code, wserr.CodeInvalidArgument)
	}

	raw, ok := err.Details["fields"]
	if !ok {
		t.Fatalf("expected fields in details")
	}

	fields, ok := raw.([]FieldError)
	if !ok {
		t.Fatalf("expected []FieldError in details, got %T", raw)
	}

	if len(fields) == 0 {
		t.Fatalf("expected non-empty fields list in details")
	}
}

func TestStruct_NilInput(t *testing.T) {
	err := Struct(nil)
	if err == nil {
		t.Fatalf("expected error for nil input")
	}

	if err.Code != wserr.CodeInvalidArgument {
		t.Fatalf("code mismatch")
	}
}
