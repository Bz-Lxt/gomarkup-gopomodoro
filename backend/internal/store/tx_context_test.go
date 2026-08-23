package store_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync/atomic"
	"testing"

	"gopomodoro/internal/store"
)

type txContextDriver struct{}

func (txContextDriver) Open(string) (driver.Conn, error) { return txContextConn{}, nil }

type txContextConn struct{}

func (txContextConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not supported") }
func (txContextConn) Close() error                        { return nil }
func (txContextConn) Begin() (driver.Tx, error)           { return txContextTx{}, nil }

type txContextTx struct{}

func (txContextTx) Commit() error   { return nil }
func (txContextTx) Rollback() error { return nil }

var txContextDriverID atomic.Uint64

func TestTxHonorsCanceledContext(t *testing.T) {
	driverName := "tx-context-" + stringID(txContextDriverID.Add(1))
	sql.Register(driverName, txContextDriver{})
	sqlDB, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	db := &store.DB{SQL: sqlDB}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err = db.Tx(ctx, func(*sql.Tx) error {
		called = true
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Tx with canceled context returned %v, want context.Canceled", err)
	}
	if called {
		t.Fatal("transaction callback ran after its context was canceled")
	}
}

func stringID(id uint64) string {
	const digits = "0123456789"
	if id == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for id > 0 {
		i--
		buf[i] = digits[id%10]
		id /= 10
	}
	return string(buf[i:])
}
