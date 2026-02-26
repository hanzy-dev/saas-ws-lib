package db

import (
	"errors"
	"testing"
	"time"
)

func TestOpenWithDriver_Validation(t *testing.T) {
	t.Parallel()

	if _, err := openWithDriver("ws_fakedb", Config{}); err == nil {
		t.Fatalf("expected error for missing DSN")
	}
	if _, err := openWithDriver("", Config{DSN: "x"}); err == nil {
		t.Fatalf("expected error for missing driver")
	}
}

func TestOpenWithDriver_PingFail(t *testing.T) {
	t.Parallel()

	st := &fakeState{pingErr: errors.New("ping fail")}
	db := openFakeDB("open_ping_fail", st)
	_ = db.Close() // ensure driver registered; openWithDriver will open again via DSN key

	_, err := openWithDriver("ws_fakedb", Config{DSN: "open_ping_fail", PingTimeout: 10 * time.Millisecond})
	if err == nil {
		t.Fatalf("expected ping error")
	}
}

func TestOpenWithDriver_PingOK_DefaultTimeout(t *testing.T) {
	t.Parallel()

	st := &fakeState{}
	db := openFakeDB("open_ok", st)
	_ = db.Close()

	got, err := openWithDriver("ws_fakedb", Config{DSN: "open_ok", PingTimeout: 0})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_ = got.Close()
}
