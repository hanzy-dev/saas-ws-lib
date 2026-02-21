package validate

import (
	"reflect"
	"strings"
	"sync"

	wserr "saas-ws-lib/pkg/errors"

	"github.com/go-playground/validator/v10"
)

var (
	once sync.Once
	v    *validator.Validate
)

func get() *validator.Validate {
	once.Do(func() {
		v = validator.New()

		// Use json tag name in validation errors (more useful for API clients)
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("json")
			if name == "" {
				return fld.Name
			}
			name = strings.Split(name, ",")[0]
			if name == "-" {
				return ""
			}
			return name
		})
	})
	return v
}

type FieldError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Param string `json:"param,omitempty"`
}

func Struct(s any) *wserr.Error {
	if s == nil {
		return wserr.New(wserr.CodeInvalidArgument, "validation failed", map[string]any{
			"fields": []FieldError{{Field: "", Tag: "nil"}},
		})
	}

	if err := get().Struct(s); err != nil {
		if ves, ok := err.(validator.ValidationErrors); ok {
			fields := make([]FieldError, 0, len(ves))
			for _, fe := range ves {
				field := fe.Field()
				if field == "" {
					field = fe.StructField()
				}
				fields = append(fields, FieldError{
					Field: field,
					Tag:   fe.Tag(),
					Param: fe.Param(),
				})
			}

			return wserr.New(wserr.CodeInvalidArgument, "validation failed", map[string]any{
				"fields": fields,
			})
		}

		return wserr.New(wserr.CodeInvalidArgument, "validation failed", map[string]any{
			"error": err.Error(),
		})
	}

	return nil
}
