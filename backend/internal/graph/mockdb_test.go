package graph_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
)

type fakeGraphDriver struct {
	mu      sync.Mutex
	handler func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error)
}

var (
	testGraphDriver = &fakeGraphDriver{}
	initOnce        sync.Once
)

func init() {
	initOnce.Do(func() {
		sql.Register("fake_graph_driver", testGraphDriver)
	})
}

func (d *fakeGraphDriver) Open(name string) (driver.Conn, error) {
	return &fakeGraphConn{driver: d}, nil
}

type fakeGraphConn struct {
	driver *fakeGraphDriver
}

func (c *fakeGraphConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeGraphStmt{conn: c, query: query}, nil
}

func (c *fakeGraphConn) Close() error {
	return nil
}

func (c *fakeGraphConn) Begin() (driver.Tx, error) {
	return &fakeGraphTx{}, nil
}

func (c *fakeGraphConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if c.driver.handler != nil {
		rows, _, err := c.driver.handler(query, args)
		return rows, err
	}
	return &fakeGraphRows{}, nil
}

func (c *fakeGraphConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if c.driver.handler != nil {
		_, res, err := c.driver.handler(query, args)
		if res == nil && err == nil {
			res = driver.RowsAffected(1)
		}
		return res, err
	}
	return driver.RowsAffected(1), nil
}

type fakeGraphStmt struct {
	conn  *fakeGraphConn
	query string
}

func (s *fakeGraphStmt) Close() error  { return nil }
func (s *fakeGraphStmt) NumInput() int { return -1 }
func (s *fakeGraphStmt) Exec(args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, a := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: a}
	}
	return s.conn.ExecContext(context.Background(), s.query, named)
}
func (s *fakeGraphStmt) Query(args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, a := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: a}
	}
	return s.conn.QueryContext(context.Background(), s.query, named)
}

type fakeGraphTx struct{}

func (t *fakeGraphTx) Commit() error   { return nil }
func (t *fakeGraphTx) Rollback() error { return nil }

type fakeGraphRows struct {
	cols []string
	data [][]driver.Value
	idx  int
}

func newRows(cols []string, data [][]driver.Value) *fakeGraphRows {
	return &fakeGraphRows{cols: cols, data: data}
}

func (r *fakeGraphRows) Columns() []string {
	return r.cols
}

func (r *fakeGraphRows) Close() error {
	return nil
}

func (r *fakeGraphRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.idx]
	copy(dest, row)
	r.idx++
	return nil
}
