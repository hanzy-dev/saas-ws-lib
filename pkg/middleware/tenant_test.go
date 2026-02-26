package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"
)

func TestTenant_FromContext_Passthrough(t *testing.T) {
	t.Parallel()

	cfg := TenantConfig{Mode: TenantFromContext, Required: true}
	h := Tenant(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wsctx.TenantID(r.Context()) != "t1" {
			t.Fatalf("tenant_id=%q", wsctx.TenantID(r.Context()))
		}
		w.WriteHeader(204)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(wsctx.WithTenantID(req.Context(), "t1"))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != 204 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestTenant_Required_UnauthenticatedWhenNoSubject(t *testing.T) {
	t.Parallel()

	cfg := TenantConfig{Mode: TenantFromContext, Required: true}
	h := Tenant(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != 401 {
		t.Fatalf("status=%d want=401 body=%s", rr.Code, rr.Body.String())
	}
}

func TestTenant_Required_ForbiddenWhenNoTenantButHasSubject(t *testing.T) {
	t.Parallel()

	cfg := TenantConfig{Mode: TenantFromContext, Required: true}
	h := Tenant(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(wsctx.WithSubjectID(req.Context(), "u1"))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != 403 {
		t.Fatalf("status=%d want=403 body=%s", rr.Code, rr.Body.String())
	}
}

func TestTenant_AllowHeader_InternalOnlyDefault(t *testing.T) {
	t.Parallel()

	cfg := TenantConfig{Mode: TenantAllowHeader, Required: true, Header: HeaderTenantID}
	h := Tenant(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))

	// no subject => header should be ignored => required => 401
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderTenantID, "t1")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != 401 {
		t.Fatalf("status=%d want=401 body=%s", rr.Code, rr.Body.String())
	}

	// with subject => header accepted => 204
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set(HeaderTenantID, "t1")
	req2 = req2.WithContext(wsctx.WithSubjectID(req2.Context(), "u1"))
	rr2 := httptest.NewRecorder()

	h.ServeHTTP(rr2, req2)
	if rr2.Code != 204 {
		t.Fatalf("status=%d want=204 body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestTenant_AllowHeaderWithoutAuth(t *testing.T) {
	t.Parallel()

	cfg := TenantConfig{
		Mode:                   TenantAllowHeader,
		Required:               true,
		AllowHeaderWithoutAuth: true,
	}
	h := Tenant(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wsctx.TenantID(r.Context()) != "t1" {
			t.Fatalf("tenant_id=%q", wsctx.TenantID(r.Context()))
		}
		w.WriteHeader(204)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderTenantID, "t1")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != 204 {
		t.Fatalf("status=%d want=204 body=%s", rr.Code, rr.Body.String())
	}
}
