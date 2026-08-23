package store

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"

	"database/sql/driver"
)

// --- fake driver for testing transaction semantics ---

type fakeDriver struct {
	mu         sync.Mutex
	beginCalls int
	lastTx     *fakeTx
}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	return &fakeConn{driver: d}, nil
}

type fakeConn struct {
	driver *fakeDriver
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (c *fakeConn) Close() error { return nil }
func (c *fakeConn) Begin() (driver.Tx, error) {
	return c.beginTx(context.Background())
}

// ConnBeginTx
func (c *fakeConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return c.beginTx(ctx)
}

func (c *fakeConn) beginTx(ctx context.Context) (driver.Tx, error) {
	c.driver.mu.Lock()
	c.driver.beginCalls++
	t := &fakeTx{}
	c.driver.lastTx = t
	c.driver.mu.Unlock()
	return t, nil
}

// SessionResetter
func (c *fakeConn) ResetSession(ctx context.Context) error { return nil }

// Pinger
func (c *fakeConn) Ping(ctx context.Context) error { return nil }

// Validator
func (c *fakeConn) IsValid() bool { return true }

type fakeTx struct {
	mu         sync.Mutex
	committed  bool
	rolledBack bool
}

func (t *fakeTx) Commit() error {
	t.mu.Lock()
	t.committed = true
	t.mu.Unlock()
	return nil
}
func (t *fakeTx) Rollback() error {
	t.mu.Lock()
	t.rolledBack = true
	t.mu.Unlock()
	return nil
}

func init() {
	sql.Register("fakefortex", &fakeDriver{})
}

func newFakeDB(t *testing.T) (*DB, *fakeDriver) {
	t.Helper()
	rawDB, err := sql.Open("fakefortex", "")
	if err != nil {
		t.Fatal(err)
	}
	drv := rawDB.Driver().(*fakeDriver)
	return &DB{SQL: rawDB}, drv
}

// --- tests ---

// When the context is already canceled before entering Tx, no transaction
// should be started, no callback should run, and nothing should commit.
func TestTxRejectsCanceledContext(t *testing.T) {
	db, drv := newFakeDB(t)
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	callbackRan := false
	err := db.Tx(ctx, func(tx *sql.Tx) error {
		callbackRan = true
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if callbackRan {
		t.Fatal("callback should not have run")
	}
	drv.mu.Lock()
	begins := drv.beginCalls
	drv.mu.Unlock()
	if begins != 0 {
		t.Fatalf("BeginTx should not have been called, got %d calls", begins)
	}
}

// When the context is canceled *during* the callback (after BeginTx but before
// Commit), the transaction must be rolled back, not committed.
func TestTxRollsBackWhenCanceledDuringCallback(t *testing.T) {
	db, drv := newFakeDB(t)
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())

	err := db.Tx(ctx, func(tx *sql.Tx) error {
		cancel() // cancel while the callback is running
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	drv.mu.Lock()
	tx := drv.lastTx
	drv.mu.Unlock()
	if tx == nil {
		t.Fatal("no transaction was created")
	}
	tx.mu.Lock()
	committed := tx.committed
	rolledBack := tx.rolledBack
	tx.mu.Unlock()
	if committed {
		t.Fatal("transaction should NOT have been committed")
	}
	if !rolledBack {
		t.Fatal("transaction should have been rolled back")
	}
}

// Normal (non-canceled) path: callback runs and commit happens.
func TestTxCommitsOnSuccess(t *testing.T) {
	db, drv := newFakeDB(t)
	defer db.Close()

	callbackRan := false
	err := db.Tx(context.Background(), func(tx *sql.Tx) error {
		callbackRan = true
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !callbackRan {
		t.Fatal("callback should have run")
	}

	drv.mu.Lock()
	tx := drv.lastTx
	drv.mu.Unlock()
	if tx == nil {
		t.Fatal("no transaction was created")
	}
	tx.mu.Lock()
	committed := tx.committed
	rolledBack := tx.rolledBack
	tx.mu.Unlock()
	if !committed {
		t.Fatal("transaction should have been committed")
	}
	if rolledBack {
		t.Fatal("transaction should NOT have been rolled back")
	}
}

// When the callback returns an error, the transaction is rolled back.
func TestTxRollsBackOnError(t *testing.T) {
	db, drv := newFakeDB(t)
	defer db.Close()

	sentinel := errors.New("callback failure")
	err := db.Tx(context.Background(), func(tx *sql.Tx) error {
		return sentinel
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	drv.mu.Lock()
	tx := drv.lastTx
	drv.mu.Unlock()
	if tx == nil {
		t.Fatal("no transaction was created")
	}
	tx.mu.Lock()
	committed := tx.committed
	rolledBack := tx.rolledBack
	tx.mu.Unlock()
	if committed {
		t.Fatal("transaction should NOT have been committed")
	}
	if !rolledBack {
		t.Fatal("transaction should have been rolled back")
	}
}
