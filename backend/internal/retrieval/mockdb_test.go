package retrieval_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
)

type fakeRetrievalDriver struct {
	mu      sync.Mutex
	handler func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error)
}

var (
	testRetrievalDriver = &fakeRetrievalDriver{}
	initOnce            sync.Once
)

func init() {
	initOnce.Do(func() {
		sql.Register("fake_retrieval_driver", testRetrievalDriver)
	})
}

func (d *fakeRetrievalDriver) Open(name string) (driver.Conn, error) {
	return &fakeRetrievalConn{driver: d}, nil
}

type fakeRetrievalConn struct {
	driver *fakeRetrievalDriver
}

func (c *fakeRetrievalConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeRetrievalStmt{conn: c, query: query}, nil
}

func (c *fakeRetrievalConn) Close() error {
	return nil
}

func (c *fakeRetrievalConn) Begin() (driver.Tx, error) {
	return &fakeRetrievalTx{}, nil
}

func (c *fakeRetrievalConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if c.driver.handler != nil {
		rows, _, err := c.driver.handler(query, args)
		return rows, err
	}
	return &fakeRetrievalRows{}, nil
}

func (c *fakeRetrievalConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
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

type fakeRetrievalStmt struct {
	conn  *fakeRetrievalConn
	query string
}

func (s *fakeRetrievalStmt) Close() error  { return nil }
func (s *fakeRetrievalStmt) NumInput() int { return -1 }
func (s *fakeRetrievalStmt) Exec(args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, a := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: a}
	}
	return s.conn.ExecContext(context.Background(), s.query, named)
}
func (s *fakeRetrievalStmt) Query(args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, a := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: a}
	}
	return s.conn.QueryContext(context.Background(), s.query, named)
}

type fakeRetrievalTx struct{}

func (t *fakeRetrievalTx) Commit() error   { return nil }
func (t *fakeRetrievalTx) Rollback() error { return nil }

type fakeRetrievalRows struct {
	cols []string
	data [][]driver.Value
	idx  int
}

func newRows(cols []string, data [][]driver.Value) *fakeRetrievalRows {
	return &fakeRetrievalRows{cols: cols, data: data}
}

func (r *fakeRetrievalRows) Columns() []string {
	return r.cols
}

func (r *fakeRetrievalRows) Close() error {
	return nil
}

func (r *fakeRetrievalRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.idx]
	copy(dest, row)
	r.idx++
	return nil
}
