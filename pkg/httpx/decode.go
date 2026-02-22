package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

func DecodeJSON(r *http.Request, dst any) *wserr.Error {
	if r == nil || dst == nil {
		return wserr.New(wserr.CodeInvalidArgument, "invalid request", nil)
	}

	if r.Body == nil || r.Body == http.NoBody {
		return wserr.New(wserr.CodeInvalidArgument, "empty request body", nil)
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return wserr.New(wserr.CodeResourceExhausted, "request body too large", map[string]any{
				"limit_bytes": maxErr.Limit,
			})
		}
		if errors.Is(err, io.EOF) {
			return wserr.New(wserr.CodeInvalidArgument, "empty request body", nil)
		}
		return wserr.New(wserr.CodeInvalidArgument, "invalid json body", nil)
	}

	if dec.More() {
		return wserr.New(wserr.CodeInvalidArgument, "invalid json body", nil)
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return wserr.New(wserr.CodeInvalidArgument, "invalid json body", nil)
	}

	return nil
}
