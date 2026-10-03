package evidence_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
)

type fakeEvidenceDriver struct {
	mu      sync.Mutex
	handler func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error)
}

var (
	testEvidenceDriver = &fakeEvidenceDriver{}
	initOnce           sync.Once
)

func init() {
	initOnce.Do(func() {
		sql.Register("fake_evidence_driver", testEvidenceDriver)
	})
}

func (d *fakeEvidenceDriver) Open(name string) (driver.Conn, error) {
	return &fakeEvidenceConn{driver: d}, nil
}

type fakeEvidenceConn struct {
	driver *fakeEvidenceDriver
}

func (c *fakeEvidenceConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeEvidenceStmt{conn: c, query: query}, nil
}

func (c *fakeEvidenceConn) Close() error {
	return nil
}

func (c *fakeEvidenceConn) Begin() (driver.Tx, error) {
	return &fakeEvidenceTx{}, nil
}

func (c *fakeEvidenceConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if c.driver.handler != nil {
		rows, _, err := c.driver.handler(query, args)
		return rows, err
	}
	return &fakeEvidenceRows{}, nil
}

func (c *fakeEvidenceConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
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

type fakeEvidenceStmt struct {
	conn  *fakeEvidenceConn
	query string
}

func (s *fakeEvidenceStmt) Close() error  { return nil }
func (s *fakeEvidenceStmt) NumInput() int { return -1 }
func (s *fakeEvidenceStmt) Exec(args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, a := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: a}
	}
	return s.conn.ExecContext(context.Background(), s.query, named)
}
func (s *fakeEvidenceStmt) Query(args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, a := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: a}
	}
	return s.conn.QueryContext(context.Background(), s.query, named)
}

type fakeEvidenceTx struct{}

func (t *fakeEvidenceTx) Commit() error   { return nil }
func (t *fakeEvidenceTx) Rollback() error { return nil }

type fakeEvidenceRows struct {
	cols []string
	data [][]driver.Value
	idx  int
}

func newRows(cols []string, data [][]driver.Value) *fakeEvidenceRows {
	return &fakeEvidenceRows{cols: cols, data: data}
}

func (r *fakeEvidenceRows) Columns() []string {
	return r.cols
}

func (r *fakeEvidenceRows) Close() error {
	return nil
}

func (r *fakeEvidenceRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.idx]
	copy(dest, row)
	r.idx++
	return nil
}
