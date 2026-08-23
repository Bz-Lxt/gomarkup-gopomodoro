package store_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"gopomodoro/internal/store"
)

var migrationDriverSequence atomic.Uint64

func TestMigrateReleasesAdvisoryLockBeforeReturningConnection(t *testing.T) {
	state := &advisoryLockState{}
	name := "migration-lock-" + stringID(migrationDriverSequence.Add(1))
	sql.Register(name, &advisoryDriver{state: state})

	migrationSQL, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	defer migrationSQL.Close()
	migrationSQL.SetMaxOpenConns(1)
	migrationSQL.SetMaxIdleConns(1)

	db := &store.DB{SQL: migrationSQL}
	if err := db.Migrate(context.Background(), t.TempDir()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	probeSQL, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	defer probeSQL.Close()

	var acquired bool
	if err := probeSQL.QueryRowContext(context.Background(), "SELECT pg_try_advisory_lock($1)", int64(88217601)).Scan(&acquired); err != nil {
		t.Fatalf("probe advisory lock: %v", err)
	}
	if !acquired {
		t.Fatal("Migrate returned while its database session still held the advisory lock")
	}
}

type advisoryLockState struct {
	mu    sync.Mutex
	owner uint64
	next  uint64
}

type advisoryDriver struct{ state *advisoryLockState }

func (d *advisoryDriver) Open(string) (driver.Conn, error) {
	d.state.mu.Lock()
	d.state.next++
	id := d.state.next
	d.state.mu.Unlock()
	return &advisoryConn{state: d.state, id: id}, nil
}

type advisoryConn struct {
	state *advisoryLockState
	id    uint64
}

func (c *advisoryConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *advisoryConn) Close() error                        { return nil }
func (c *advisoryConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }

func (c *advisoryConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	switch {
	case strings.Contains(query, "pg_advisory_lock"):
		c.state.owner = c.id
	case strings.Contains(query, "pg_advisory_unlock") && c.state.owner == c.id:
		c.state.owner = 0
	}
	return driver.RowsAffected(0), nil
}

func (c *advisoryConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	if strings.Contains(query, "pg_try_advisory_lock") {
		acquired := c.state.owner == 0 || c.state.owner == c.id
		if acquired {
			c.state.owner = c.id
		}
		return &singleBoolRow{value: acquired}, nil
	}
	return nil, driver.ErrSkip
}

type singleBoolRow struct {
	value bool
	done  bool
}

func (r *singleBoolRow) Columns() []string { return []string{"pg_try_advisory_lock"} }
func (r *singleBoolRow) Close() error      { return nil }
func (r *singleBoolRow) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.value
	return nil
}

func stringID(n uint64) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = digits[n%10]
		n /= 10
	}
	return string(buf[i:])
}
