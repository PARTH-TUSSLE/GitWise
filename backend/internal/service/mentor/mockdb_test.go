package mentor_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
)

type fakeMentorDriver struct {
	mu      sync.Mutex
	handler func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error)
}

var (
	testMentorDriver = &fakeMentorDriver{}
	initOnce         sync.Once
)

func init() {
	initOnce.Do(func() {
		sql.Register("fake_mentor_driver", testMentorDriver)
	})
}

func (d *fakeMentorDriver) Open(name string) (driver.Conn, error) {
	return &fakeMentorConn{driver: d}, nil
}

type fakeMentorConn struct {
	driver *fakeMentorDriver
}

func (c *fakeMentorConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeMentorStmt{conn: c, query: query}, nil
}

func (c *fakeMentorConn) Close() error {
	return nil
}

func (c *fakeMentorConn) Begin() (driver.Tx, error) {
	return &fakeMentorTx{}, nil
}

func (c *fakeMentorConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if c.driver.handler != nil {
		rows, _, err := c.driver.handler(query, args)
		return rows, err
	}
	return &fakeMentorRows{}, nil
}

func (c *fakeMentorConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
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

type fakeMentorStmt struct {
	conn  *fakeMentorConn
	query string
}

func (s *fakeMentorStmt) Close() error  { return nil }
func (s *fakeMentorStmt) NumInput() int { return -1 }
func (s *fakeMentorStmt) Exec(args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, a := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: a}
	}
	return s.conn.ExecContext(context.Background(), s.query, named)
}
func (s *fakeMentorStmt) Query(args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, a := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: a}
	}
	return s.conn.QueryContext(context.Background(), s.query, named)
}

type fakeMentorTx struct{}

func (t *fakeMentorTx) Commit() error   { return nil }
func (t *fakeMentorTx) Rollback() error { return nil }

type fakeMentorRows struct {
	cols []string
	data [][]driver.Value
	idx  int
}

func newRows(cols []string, data [][]driver.Value) *fakeMentorRows {
	return &fakeMentorRows{cols: cols, data: data}
}

func (r *fakeMentorRows) Columns() []string {
	return r.cols
}

func (r *fakeMentorRows) Close() error {
	return nil
}

func (r *fakeMentorRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.idx]
	copy(dest, row)
	r.idx++
	return nil
}
