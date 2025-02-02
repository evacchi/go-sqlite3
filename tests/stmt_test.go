package tests

import (
	"math"
	"testing"
	"time"

	"github.com/ncruces/go-sqlite3"
)

func TestStmt(t *testing.T) {
	t.Parallel()

	db, err := sqlite3.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Exec(`CREATE TABLE IF NOT EXISTS test (col)`)
	if err != nil {
		t.Fatal(err)
	}

	stmt, _, err := db.Prepare(`INSERT INTO test(col) VALUES(?)`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	if got := stmt.BindCount(); got != 1 {
		t.Errorf("got %d, want 1", got)
	}

	if err := stmt.BindBool(1, false); err != nil {
		t.Fatal(err)
	}
	if err := stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	if err := stmt.BindBool(1, true); err != nil {
		t.Fatal(err)
	}
	if err := stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	if err := stmt.BindInt(1, 2); err != nil {
		t.Fatal(err)
	}
	if err = stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	if err := stmt.BindFloat(1, math.Pi); err != nil {
		t.Fatal(err)
	}
	if err := stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	if err := stmt.BindNull(1); err != nil {
		t.Fatal(err)
	}
	if err := stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	if err := stmt.BindText(1, ""); err != nil {
		t.Fatal(err)
	}
	if err := stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	if err := stmt.BindText(1, "text"); err != nil {
		t.Fatal(err)
	}
	if err := stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	if err := stmt.BindBlob(1, []byte("blob")); err != nil {
		t.Fatal(err)
	}
	if err := stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	if err := stmt.BindBlob(1, nil); err != nil {
		t.Fatal(err)
	}
	if err := stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	if err := stmt.BindZeroBlob(1, 4); err != nil {
		t.Fatal(err)
	}
	if err := stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	if err := stmt.ClearBindings(); err != nil {
		t.Fatal(err)
	}
	if err := stmt.Exec(); err != nil {
		t.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		t.Fatal(err)
	}

	// The table should have: 0, 1, 2, π, NULL, "", "text", "blob", NULL, "\0\0\0\0", NULL
	stmt, _, err = db.Prepare(`SELECT col FROM test`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.INTEGER {
			t.Errorf("got %v, want INTEGER", got)
		}
		if got := stmt.ColumnBool(0); got != false {
			t.Errorf("got %v, want false", got)
		}
		if got := stmt.ColumnInt(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnFloat(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnText(0); got != "0" {
			t.Errorf("got %q, want zero", got)
		}
		if got := stmt.ColumnBlob(0, nil); string(got) != "0" {
			t.Errorf("got %q, want zero", got)
		}
	}

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.INTEGER {
			t.Errorf("got %v, want INTEGER", got)
		}
		if got := stmt.ColumnBool(0); got != true {
			t.Errorf("got %v, want true", got)
		}
		if got := stmt.ColumnInt(0); got != 1 {
			t.Errorf("got %v, want one", got)
		}
		if got := stmt.ColumnFloat(0); got != 1 {
			t.Errorf("got %v, want one", got)
		}
		if got := stmt.ColumnText(0); got != "1" {
			t.Errorf("got %q, want one", got)
		}
		if got := stmt.ColumnBlob(0, nil); string(got) != "1" {
			t.Errorf("got %q, want one", got)
		}
	}

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.INTEGER {
			t.Errorf("got %v, want INTEGER", got)
		}
		if got := stmt.ColumnBool(0); got != true {
			t.Errorf("got %v, want true", got)
		}
		if got := stmt.ColumnInt(0); got != 2 {
			t.Errorf("got %v, want two", got)
		}
		if got := stmt.ColumnFloat(0); got != 2 {
			t.Errorf("got %v, want two", got)
		}
		if got := stmt.ColumnText(0); got != "2" {
			t.Errorf("got %q, want two", got)
		}
		if got := stmt.ColumnBlob(0, nil); string(got) != "2" {
			t.Errorf("got %q, want two", got)
		}
	}

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.FLOAT {
			t.Errorf("got %v, want FLOAT", got)
		}
		if got := stmt.ColumnBool(0); got != true {
			t.Errorf("got %v, want true", got)
		}
		if got := stmt.ColumnInt(0); got != 3 {
			t.Errorf("got %v, want three", got)
		}
		if got := stmt.ColumnFloat(0); got != math.Pi {
			t.Errorf("got %v, want π", got)
		}
		if got := stmt.ColumnText(0); got != "3.14159265358979" {
			t.Errorf("got %q, want π", got)
		}
		if got := stmt.ColumnBlob(0, nil); string(got) != "3.14159265358979" {
			t.Errorf("got %q, want π", got)
		}
	}

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.NULL {
			t.Errorf("got %v, want NULL", got)
		}
		if got := stmt.ColumnBool(0); got != false {
			t.Errorf("got %v, want false", got)
		}
		if got := stmt.ColumnInt(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnFloat(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnText(0); got != "" {
			t.Errorf("got %q, want empty", got)
		}
		if got := stmt.ColumnBlob(0, nil); got != nil {
			t.Errorf("got %q, want nil", got)
		}
	}

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.TEXT {
			t.Errorf("got %v, want TEXT", got)
		}
		if got := stmt.ColumnBool(0); got != false {
			t.Errorf("got %v, want false", got)
		}
		if got := stmt.ColumnInt(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnFloat(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnText(0); got != "" {
			t.Errorf("got %q, want empty", got)
		}
		if got := stmt.ColumnBlob(0, nil); got != nil {
			t.Errorf("got %q, want nil", got)
		}
	}

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.TEXT {
			t.Errorf("got %v, want TEXT", got)
		}
		if got := stmt.ColumnBool(0); got != false {
			t.Errorf("got %v, want false", got)
		}
		if got := stmt.ColumnInt(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnFloat(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnText(0); got != "text" {
			t.Errorf(`got %q, want "text"`, got)
		}
		if got := stmt.ColumnBlob(0, nil); string(got) != "text" {
			t.Errorf(`got %q, want "text"`, got)
		}
	}

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.BLOB {
			t.Errorf("got %v, want BLOB", got)
		}
		if got := stmt.ColumnBool(0); got != false {
			t.Errorf("got %v, want false", got)
		}
		if got := stmt.ColumnInt(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnFloat(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnText(0); got != "blob" {
			t.Errorf(`got %q, want "blob"`, got)
		}
		if got := stmt.ColumnBlob(0, nil); string(got) != "blob" {
			t.Errorf(`got %q, want "blob"`, got)
		}
	}

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.NULL {
			t.Errorf("got %v, want NULL", got)
		}
		if got := stmt.ColumnBool(0); got != false {
			t.Errorf("got %v, want false", got)
		}
		if got := stmt.ColumnInt(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnFloat(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnText(0); got != "" {
			t.Errorf("got %q, want empty", got)
		}
		if got := stmt.ColumnBlob(0, nil); got != nil {
			t.Errorf("got %q, want nil", got)
		}
	}

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.BLOB {
			t.Errorf("got %v, want BLOB", got)
		}
		if got := stmt.ColumnBool(0); got != false {
			t.Errorf("got %v, want false", got)
		}
		if got := stmt.ColumnInt(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnFloat(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnText(0); got != "\x00\x00\x00\x00" {
			t.Errorf(`got %q, want "\x00\x00\x00\x00"`, got)
		}
		if got := stmt.ColumnBlob(0, nil); string(got) != "\x00\x00\x00\x00" {
			t.Errorf(`got %q, want "\x00\x00\x00\x00"`, got)
		}
	}

	if stmt.Step() {
		if got := stmt.ColumnType(0); got != sqlite3.NULL {
			t.Errorf("got %v, want NULL", got)
		}
		if got := stmt.ColumnBool(0); got != false {
			t.Errorf("got %v, want false", got)
		}
		if got := stmt.ColumnInt(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnFloat(0); got != 0 {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnText(0); got != "" {
			t.Errorf("got %q, want empty", got)
		}
		if got := stmt.ColumnBlob(0, nil); got != nil {
			t.Errorf("got %q, want nil", got)
		}
	}

	if err := stmt.Close(); err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestStmt_Close(t *testing.T) {
	var stmt *sqlite3.Stmt
	stmt.Close()
}

func TestStmt_BindName(t *testing.T) {
	t.Parallel()

	db, err := sqlite3.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	want := []string{"", "", "", "", "?5", ":AAA", "@AAA", "$AAA"}
	stmt, _, err := db.Prepare(`SELECT ?, ?5, :AAA, @AAA, $AAA`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	if got := stmt.BindCount(); got != len(want) {
		t.Errorf("got %d, want %d", got, len(want))
	}

	for i, name := range want {
		id := i + 1
		if got := stmt.BindName(id); got != name {
			t.Errorf("got %q, want %q", got, name)
		}
		if name == "" {
			id = 0
		}
		if got := stmt.BindIndex(name); got != id {
			t.Errorf("got %d, want %d", got, id)
		}
	}
}

func TestStmt_ColumnTime(t *testing.T) {
	t.Parallel()

	db, err := sqlite3.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	stmt, _, err := db.Prepare(`SELECT ?, ?, ?, datetime(), unixepoch(), julianday(), NULL, 'abc'`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	reference := time.Date(2013, 10, 7, 4, 23, 19, 120_000_000, time.FixedZone("", -4*3600))
	err = stmt.BindTime(1, reference, sqlite3.TimeFormat4)
	if err != nil {
		t.Fatal(err)
	}
	err = stmt.BindTime(2, reference, sqlite3.TimeFormatUnixMilli)
	if err != nil {
		t.Fatal(err)
	}
	err = stmt.BindTime(3, reference, sqlite3.TimeFormatJulianDay)
	if err != nil {
		t.Fatal(err)
	}

	if now := time.Now(); stmt.Step() {
		if got := stmt.ColumnTime(0, sqlite3.TimeFormatAuto); !got.Equal(reference) {
			t.Errorf("got %v, want %v", got, reference)
		}
		if got := stmt.ColumnTime(1, sqlite3.TimeFormatAuto); !got.Equal(reference) {
			t.Errorf("got %v, want %v", got, reference)
		}
		if got := stmt.ColumnTime(2, sqlite3.TimeFormatAuto); got.Sub(reference).Abs() > time.Millisecond {
			t.Errorf("got %v, want %v", got, reference)
		}

		if got := stmt.ColumnTime(3, sqlite3.TimeFormatAuto); got.Sub(now).Abs() > time.Second {
			t.Errorf("got %v, want %v", got, now)
		}
		if got := stmt.ColumnTime(4, sqlite3.TimeFormatAuto); got.Sub(now).Abs() > time.Second {
			t.Errorf("got %v, want %v", got, now)
		}
		if got := stmt.ColumnTime(5, sqlite3.TimeFormatAuto); got.Sub(now).Abs() > time.Second/10 {
			t.Errorf("got %v, want %v", got, now)
		}

		if got := stmt.ColumnTime(6, sqlite3.TimeFormatAuto); got != (time.Time{}) {
			t.Errorf("got %v, want zero", got)
		}
		if got := stmt.ColumnTime(7, sqlite3.TimeFormatAuto); got != (time.Time{}) {
			t.Errorf("got %v, want zero", got)
		}
		if stmt.Err() == nil {
			t.Errorf("want error")
		}
	}
}
