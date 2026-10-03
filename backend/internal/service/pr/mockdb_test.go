package pr_test

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
)

func init() {
	sql.Register("fake_pr_driver", &fakePRDriver{})
}

type fakePRDriver struct{}

func (d *fakePRDriver) Open(name string) (driver.Conn, error) {
	return &fakePRConn{}, nil
}

type fakePRConn struct{}

func (c *fakePRConn) Prepare(query string) (driver.Stmt, error) {
	return &fakePRStmt{query: query}, nil
}

func (c *fakePRConn) Close() error {
	return nil
}

func (c *fakePRConn) Begin() (driver.Tx, error) {
	return &fakePRTx{}, nil
}

func (c *fakePRConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return c.ExecContext(nil, query, named)
}

func (c *fakePRConn) ExecContext(ctx interface{}, query string, args []driver.NamedValue) (driver.Result, error) {
	testPRDriver.mu.Lock()
	defer testPRDriver.mu.Unlock()
	if testPRDriver.handler != nil {
		_, res, err := testPRDriver.handler(query, args)
		return res, err
	}
	return driver.RowsAffected(1), nil
}

func (c *fakePRConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return c.QueryContext(nil, query, named)
}

func (c *fakePRConn) QueryContext(ctx interface{}, query string, args []driver.NamedValue) (driver.Rows, error) {
	testPRDriver.mu.Lock()
	defer testPRDriver.mu.Unlock()
	if testPRDriver.handler != nil {
		rows, _, err := testPRDriver.handler(query, args)
		return rows, err
	}
	return newRows([]string{}, nil), nil
}

type fakePRStmt struct {
	query string
}

func (s *fakePRStmt) Close() error {
	return nil
}

func (s *fakePRStmt) NumInput() int {
	return -1
}

func (s *fakePRStmt) Exec(args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	testPRDriver.mu.Lock()
	defer testPRDriver.mu.Unlock()
	if testPRDriver.handler != nil {
		_, res, err := testPRDriver.handler(s.query, named)
		return res, err
	}
	return driver.RowsAffected(1), nil
}

func (s *fakePRStmt) Query(args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	testPRDriver.mu.Lock()
	defer testPRDriver.mu.Unlock()
	if testPRDriver.handler != nil {
		rows, _, err := testPRDriver.handler(s.query, named)
		return rows, err
	}
	return newRows([]string{}, nil), nil
}

type fakePRTx struct{}

func (t *fakePRTx) Commit() error   { return nil }
func (t *fakePRTx) Rollback() error { return nil }

type prDriverTracker struct {
	mu      sync.Mutex
	handler func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error)
}

var testPRDriver prDriverTracker

type fakeRows struct {
	columns []string
	data    [][]driver.Value
	idx     int
}

func newRows(columns []string, data [][]driver.Value) driver.Rows {
	return &fakeRows{
		columns: columns,
		data:    data,
		idx:     0,
	}
}

func (r *fakeRows) Columns() []string {
	return r.columns
}

func (r *fakeRows) Close() error {
	return nil
}

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.idx]
	copy(dest, row)
	r.idx++
	return nil
}
