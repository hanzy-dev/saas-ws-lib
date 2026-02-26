package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

type fakeState struct {
	mu sync.Mutex

	version int64

	// tx signals
	beginErr    error
	commitErr   error
	rollbackErr error

	// exec/query/ping
	execErr  error
	queryErr error
	pingErr  error

	commits   int
	rollbacks int
	execs     int
	queries   int
}

var (
	fakeOnce sync.Once
	fakeMu   sync.Mutex
	fakeDBs  = map[string]*fakeState{}
)

func registerFakeDriver() {
	fakeOnce.Do(func() {
		sql.Register("ws_fakedb", fakeDriver{})
	})
}

func openFakeDB(tName string, st *fakeState) *sql.DB {
	registerFakeDriver()

	fakeMu.Lock()
	fakeDBs[tName] = st
	fakeMu.Unlock()

	db, err := sql.Open("ws_fakedb", tName)
	if err != nil {
		panic(err)
	}
	return db
}

type fakeDriver struct{}

func (fakeDriver) Open(name string) (driver.Conn, error) {
	fakeMu.Lock()
	st := fakeDBs[name]
	fakeMu.Unlock()
	if st == nil {
		return nil, fmt.Errorf("unknown fakedb name: %s", name)
	}
	return &fakeConn{st: st}, nil
}

type fakeConn struct {
	st *fakeState
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not supported")
}
func (c *fakeConn) Close() error { return nil }
func (c *fakeConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *fakeConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	c.st.mu.Lock()
	defer c.st.mu.Unlock()
	if c.st.beginErr != nil {
		return nil, c.st.beginErr
	}
	return &fakeTx{st: c.st}, nil
}

func (c *fakeConn) Ping(ctx context.Context) error {
	c.st.mu.Lock()
	defer c.st.mu.Unlock()
	return c.st.pingErr
}

func (c *fakeConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.st.mu.Lock()
	defer c.st.mu.Unlock()
	c.st.execs++
	if c.st.execErr != nil {
		return nil, c.st.execErr
	}

	// Handle UpdateSchemaVersion: first arg is version
	// Query string is ignored (we're unit testing behavior, not SQL parsing).
	if len(args) >= 1 {
		if v, ok := args[0].Value.(int64); ok {
			c.st.version = v
		}
	}
	return driver.RowsAffected(1), nil
}

func (c *fakeConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.st.mu.Lock()
	defer c.st.mu.Unlock()
	c.st.queries++
	if c.st.queryErr != nil {
		return nil, c.st.queryErr
	}
	return &fakeRows{cols: []string{"version"}, vals: [][]any{{c.st.version}}}, nil
}

var _ driver.Conn = (*fakeConn)(nil)
var _ driver.ConnBeginTx = (*fakeConn)(nil)
var _ driver.ExecerContext = (*fakeConn)(nil)
var _ driver.QueryerContext = (*fakeConn)(nil)
var _ driver.Pinger = (*fakeConn)(nil)

type fakeTx struct {
	st *fakeState
}

func (t *fakeTx) Commit() error {
	t.st.mu.Lock()
	defer t.st.mu.Unlock()
	t.st.commits++
	return t.st.commitErr
}

func (t *fakeTx) Rollback() error {
	t.st.mu.Lock()
	defer t.st.mu.Unlock()
	t.st.rollbacks++
	return t.st.rollbackErr
}

var _ driver.Tx = (*fakeTx)(nil)

type fakeRows struct {
	cols []string
	vals [][]any
	i    int
}

func (r *fakeRows) Columns() []string { return r.cols }
func (r *fakeRows) Close() error      { return nil }

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.i >= len(r.vals) {
		return io.EOF
	}
	row := r.vals[r.i]
	r.i++
	for i := range dest {
		dest[i] = row[i]
	}
	return nil
}

// just to silence unused imports when file is compiled alone
var _ = time.Now
