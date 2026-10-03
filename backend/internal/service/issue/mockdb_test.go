package issue_test

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
)

func init() {
	sql.Register("fake_issue_driver", &fakeIssueDriver{})
}

type fakeIssueDriver struct{}

func (d *fakeIssueDriver) Open(name string) (driver.Conn, error) {
	return &fakeIssueConn{}, nil
}

type fakeIssueConn struct{}

func (c *fakeIssueConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeIssueStmt{query: query}, nil
}

func (c *fakeIssueConn) Close() error {
	return nil
}

func (c *fakeIssueConn) Begin() (driver.Tx, error) {
	return &fakeIssueTx{}, nil
}

func (c *fakeIssueConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return c.ExecContext(nil, query, named)
}

func (c *fakeIssueConn) ExecContext(ctx interface{}, query string, args []driver.NamedValue) (driver.Result, error) {
	testIssueDriver.mu.Lock()
	defer testIssueDriver.mu.Unlock()
	if testIssueDriver.handler != nil {
		_, res, err := testIssueDriver.handler(query, args)
		return res, err
	}
	return driver.RowsAffected(1), nil
}

func (c *fakeIssueConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return c.QueryContext(nil, query, named)
}

func (c *fakeIssueConn) QueryContext(ctx interface{}, query string, args []driver.NamedValue) (driver.Rows, error) {
	testIssueDriver.mu.Lock()
	defer testIssueDriver.mu.Unlock()
	if testIssueDriver.handler != nil {
		rows, _, err := testIssueDriver.handler(query, args)
		return rows, err
	}
	return newRows([]string{}, nil), nil
}

type fakeIssueStmt struct {
	query string
}

func (s *fakeIssueStmt) Close() error {
	return nil
}

func (s *fakeIssueStmt) NumInput() int {
	return -1
}

func (s *fakeIssueStmt) Exec(args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	testIssueDriver.mu.Lock()
	defer testIssueDriver.mu.Unlock()
	if testIssueDriver.handler != nil {
		_, res, err := testIssueDriver.handler(s.query, named)
		return res, err
	}
	return driver.RowsAffected(1), nil
}

func (s *fakeIssueStmt) Query(args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	testIssueDriver.mu.Lock()
	defer testIssueDriver.mu.Unlock()
	if testIssueDriver.handler != nil {
		rows, _, err := testIssueDriver.handler(s.query, named)
		return rows, err
	}
	return newRows([]string{}, nil), nil
}

type fakeIssueTx struct{}

func (t *fakeIssueTx) Commit() error   { return nil }
func (t *fakeIssueTx) Rollback() error { return nil }

type driverTracker struct {
	mu      sync.Mutex
	handler func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error)
}

var testIssueDriver driverTracker

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
