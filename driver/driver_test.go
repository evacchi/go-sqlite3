// Package driver provides a database/sql driver for SQLite.
package driver

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/ncruces/go-sqlite3"
)

func Test_Open_dir(t *testing.T) {
	db, err := sql.Open("sqlite3", ".")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Conn(context.TODO())
	if err == nil {
		t.Fatal("want error")
	}
	var serr *sqlite3.Error
	if !errors.As(err, &serr) {
		t.Fatalf("got %T, want sqlite3.Error", err)
	}
	if rc := serr.Code(); rc != sqlite3.CANTOPEN {
		t.Errorf("got %d, want sqlite3.CANTOPEN", rc)
	}
	if got := err.Error(); got != `sqlite3: unable to open database file` {
		t.Error("got message: ", got)
	}
}

func Test_Open_pragma(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:?_pragma=busy_timeout(1000)")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var timeout int
	err = db.QueryRow(`PRAGMA busy_timeout`).Scan(&timeout)
	if err != nil {
		t.Fatal(err)
	}
	if timeout != 1000 {
		t.Errorf("got %v, want 1000", timeout)
	}
}

func Test_Open_pragma_invalid(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:?_pragma=busy_timeout+1000")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Conn(context.TODO())
	if err == nil {
		t.Fatal("want error")
	}
	var serr *sqlite3.Error
	if !errors.As(err, &serr) {
		t.Fatalf("got %T, want sqlite3.Error", err)
	}
	if rc := serr.Code(); rc != sqlite3.ERROR {
		t.Errorf("got %d, want sqlite3.ERROR", rc)
	}
	if got := err.Error(); got != `sqlite3: invalid _pragma: sqlite3: SQL logic error: near "1000": syntax error` {
		t.Error("got message: ", got)
	}
}

func Test_Open_txLock(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:"+
		filepath.Join(t.TempDir(), "test.db")+
		"?_txlock=exclusive&_pragma=busy_timeout(0)")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tx1, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Begin()
	if err == nil {
		t.Error("want error")
	}
	var serr *sqlite3.Error
	if !errors.As(err, &serr) {
		t.Fatalf("got %T, want sqlite3.Error", err)
	}
	if rc := serr.Code(); rc != sqlite3.BUSY {
		t.Errorf("got %d, want sqlite3.BUSY", rc)
	}
	var terr interface{ Temporary() bool }
	if !errors.As(err, &terr) || !terr.Temporary() {
		t.Error("not temporary", err)
	}
	if got := err.Error(); got != `sqlite3: database is locked` {
		t.Error("got message: ", got)
	}

	err = tx1.Commit()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Open_txLock_invalid(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:?_txlock=xclusive")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Conn(context.TODO())
	if err == nil {
		t.Fatal("want error")
	}
	if got := err.Error(); got != `sqlite3: invalid _txlock: xclusive` {
		t.Error("got message: ", got)
	}
}

func Test_BeginTx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err.Error() != string(isolationErr) {
		t.Error("want isolationErr")
	}

	tx1, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	tx2, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx1.Exec(`CREATE TABLE IF NOT EXISTS test (col)`)
	if err == nil {
		t.Error("want error")
	}
	var serr *sqlite3.Error
	if !errors.As(err, &serr) {
		t.Fatalf("got %T, want sqlite3.Error", err)
	}
	if rc := serr.Code(); rc != sqlite3.READONLY {
		t.Errorf("got %d, want sqlite3.READONLY", rc)
	}
	if got := err.Error(); got != `sqlite3: attempt to write a readonly database` {
		t.Error("got message: ", got)
	}

	err = tx2.Commit()
	if err != nil {
		t.Fatal(err)
	}

	err = tx1.Commit()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Prepare(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare(`SELECT 1; -- HERE`)
	if err != nil {
		t.Error(err)
	}
	defer stmt.Close()

	var serr *sqlite3.Error
	_, err = db.Prepare(`SELECT`)
	if err == nil {
		t.Error("want error")
	}
	if !errors.As(err, &serr) {
		t.Fatalf("got %T, want sqlite3.Error", err)
	}
	if rc := serr.Code(); rc != sqlite3.ERROR {
		t.Errorf("got %d, want sqlite3.ERROR", rc)
	}
	if got := err.Error(); got != `sqlite3: SQL logic error: incomplete input` {
		t.Error("got message: ", got)
	}

	_, err = db.Prepare(`SELECT 1; SELECT`)
	if err == nil {
		t.Error("want error")
	}
	if !errors.As(err, &serr) {
		t.Fatalf("got %T, want sqlite3.Error", err)
	}
	if rc := serr.Code(); rc != sqlite3.ERROR {
		t.Errorf("got %d, want sqlite3.ERROR", rc)
	}
	if got := err.Error(); got != `sqlite3: SQL logic error: incomplete input` {
		t.Error("got message: ", got)
	}

	_, err = db.Prepare(`SELECT 1; SELECT 2`)
	if err.Error() != string(tailErr) {
		t.Error("want tailErr")
	}
}

func Test_QueryRow_named(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	stmt, err := conn.PrepareContext(ctx, `SELECT ?, ?5, :AAA, @AAA, $AAA`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	date := time.Now()
	row := stmt.QueryRow(true, sql.Named("AAA", math.Pi), nil /*3*/, nil /*4*/, date /*5*/)

	var first bool
	var fifth time.Time
	var colon, at, dollar float32
	err = row.Scan(&first, &fifth, &colon, &at, &dollar)
	if err != nil {
		t.Fatal(err)
	}

	if first != true {
		t.Errorf("want true, got %v", first)
	}
	if colon != math.Pi {
		t.Errorf("want π, got %v", colon)
	}
	if at != math.Pi {
		t.Errorf("want π, got %v", at)
	}
	if dollar != math.Pi {
		t.Errorf("want π, got %v", dollar)
	}
	if !fifth.Equal(date) {
		t.Errorf("want %v, got %v", date, fifth)
	}
}

func Test_QueryRow_blob_null(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT NULL    UNION ALL
		SELECT x'cafe' UNION ALL
		SELECT x'babe' UNION ALL
		SELECT NULL
	`)
	if err != nil {
		t.Fatal(err)
	}

	want := [][]byte{nil, {0xca, 0xfe}, {0xba, 0xbe}, nil}
	for i := 0; rows.Next(); i++ {
		var buf []byte
		err = rows.Scan(&buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, want[i]) {
			t.Errorf("got %q, want %q", buf, want[i])
		}
	}
}

func Test_ZeroBlob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS test (col)`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.ExecContext(ctx, `INSERT INTO test(col) VALUES(?)`, sqlite3.ZeroBlob(4))
	if err != nil {
		t.Fatal(err)
	}

	var got []byte
	err = conn.QueryRowContext(ctx, `SELECT col FROM test`).Scan(&got)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "\x00\x00\x00\x00" {
		t.Errorf(`got %q, want "\x00\x00\x00\x00"`, got)
	}
}
