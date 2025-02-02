package tests

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func TestParallel(t *testing.T) {
	name := filepath.Join(t.TempDir(), "test.db")
	testParallel(t, name, 1000)
	testIntegrity(t, name)
}

func TestMultiProcess(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	name := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("TestMultiProcess_dbname", name)

	cmd := exec.Command("go", "test", "-v", "-run", "TestChildProcess")
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	var buf [3]byte
	// Wait for child to start.
	if _, err := io.ReadFull(out, buf[:]); err != nil || string(buf[:]) != "===" {
		t.Fatal(err)
	}

	testParallel(t, name, 1000)
	if err := cmd.Wait(); err != nil {
		t.Error(err)
	}
	testIntegrity(t, name)
}

func TestChildProcess(t *testing.T) {
	name := os.Getenv("TestMultiProcess_dbname")
	if name == "" || testing.Short() {
		t.SkipNow()
	}

	testParallel(t, name, 1000)
}

func testParallel(t *testing.T, name string, n int) {
	writer := func() error {
		db, err := sqlite3.Open(name)
		if err != nil {
			return err
		}
		defer db.Close()

		err = db.Exec(`
			PRAGMA busy_timeout=10000;
			PRAGMA synchronous=off;
			PRAGMA locking_mode=normal;
			PRAGMA journal_mode=truncate;
		`)
		if err != nil {
			return err
		}

		err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id INT, name VARCHAR(10))`)
		if err != nil {
			return err
		}

		err = db.Exec(`INSERT INTO users(id, name) VALUES(0, 'go'), (1, 'zig'), (2, 'whatever')`)
		if err != nil {
			return err
		}

		return db.Close()
	}

	reader := func() error {
		db, err := sqlite3.Open(name)
		if err != nil {
			return err
		}
		defer db.Close()

		err = db.Exec(`
			PRAGMA busy_timeout=10000;
			PRAGMA locking_mode=normal;
		`)
		if err != nil {
			return err
		}

		stmt, _, err := db.Prepare(`SELECT id, name FROM users`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		row := 0
		for stmt.Step() {
			row++
		}
		if err := stmt.Err(); err != nil {
			return err
		}
		if row%3 != 0 {
			t.Errorf("got %d rows, want multiple of 3", row)
		}

		err = stmt.Close()
		if err != nil {
			return err
		}

		return db.Close()
	}

	err := writer()
	if err != nil {
		t.Fatal(err)
	}

	var group errgroup.Group
	group.SetLimit(4)
	for i := 0; i < n; i++ {
		if i&7 != 7 {
			group.Go(reader)
		} else {
			group.Go(writer)
		}
	}
	err = group.Wait()
	if err != nil {
		t.Error(err)
	}
}

func testIntegrity(t *testing.T, name string) {
	db, err := sqlite3.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	test := `PRAGMA integrity_check`
	if testing.Short() {
		test = `PRAGMA quick_check`
	}

	stmt, _, err := db.Prepare(test)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	for stmt.Step() {
		if row := stmt.ColumnText(0); row != "ok" {
			t.Error(row)
		}
	}
	if err := stmt.Err(); err != nil {
		t.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}
