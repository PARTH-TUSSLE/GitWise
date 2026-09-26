package repo_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
)

type fakeRepoDriver struct {
	mu      sync.Mutex
	handler func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error)
}

var (
	testRepoDriver = &fakeRepoDriver{}
	repoDriverOnce sync.Once
)

func init() {
	repoDriverOnce.Do(func() {
		sql.Register("fake_repo_driver", testRepoDriver)
	})
}

func (d *fakeRepoDriver) Open(name string) (driver.Conn, error) {
	return &fakeRepoConn{driver: d}, nil
}

type fakeRepoConn struct {
	driver *fakeRepoDriver
}

func (c *fakeRepoConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeRepoStmt{conn: c, query: query}, nil
}

func (c *fakeRepoConn) Close() error {
	return nil
}

func (c *fakeRepoConn) Begin() (driver.Tx, error) {
	return &fakeRepoTx{}, nil
}

func (c *fakeRepoConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if c.driver.handler != nil {
		rows, _, err := c.driver.handler(query, args)
		if rows == nil && err == nil {
			return &fakeRepoRows{}, nil
		}
		return rows, err
	}
	return &fakeRepoRows{}, nil
}

func (c *fakeRepoConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
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

type fakeRepoStmt struct {
	conn  *fakeRepoConn
	query string
}

func (s *fakeRepoStmt) Close() error  { return nil }
func (s *fakeRepoStmt) NumInput() int { return -1 }
func (s *fakeRepoStmt) Exec(args []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}
func (s *fakeRepoStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &fakeRepoRows{}, nil
}
func (s *fakeRepoStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return s.conn.QueryContext(ctx, s.query, args)
}
func (s *fakeRepoStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return s.conn.ExecContext(ctx, s.query, args)
}

type fakeRepoTx struct{}

func (t *fakeRepoTx) Commit() error   { return nil }
func (t *fakeRepoTx) Rollback() error { return nil }

type fakeRepoRows struct {
	cols []string
	data [][]driver.Value
	idx  int
}

func (r *fakeRepoRows) Columns() []string { return r.cols }
func (r *fakeRepoRows) Close() error      { return nil }
func (r *fakeRepoRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.idx]
	for i, val := range row {
		dest[i] = val
	}
	r.idx++
	return nil
}
