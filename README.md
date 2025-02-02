# Go bindings to SQLite using Wazero

[![Go Reference](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/ncruces/go-sqlite3)
[![Go Report](https://goreportcard.com/badge/github.com/ncruces/go-sqlite3)](https://goreportcard.com/report/github.com/ncruces/go-sqlite3)
[![Go Coverage](https://github.com/ncruces/go-sqlite3/wiki/coverage.svg)](https://raw.githack.com/wiki/ncruces/go-sqlite3/coverage.html)

### ⚠️ Work in Progress ⚠️

Go module `github.com/ncruces/go-sqlite3` wraps a [WASM](https://webassembly.org/) build of [SQLite](https://sqlite.org/),
and uses [wazero](https://wazero.io/) to provide `cgo`-free SQLite bindings.

- Package [`github.com/ncruces/go-sqlite3`](https://pkg.go.dev/github.com/ncruces/go-sqlite3)
wraps the [C SQLite API](https://www.sqlite.org/cintro.html)
([example usage](https://pkg.go.dev/github.com/ncruces/go-sqlite3#example-package)).
- Package [`github.com/ncruces/go-sqlite3/driver`](https://pkg.go.dev/github.com/ncruces/go-sqlite3/driver)
provides a [`database/sql`](https://pkg.go.dev/database/sql) driver
([example usage](https://pkg.go.dev/github.com/ncruces/go-sqlite3/driver#example-package)).
- Package [`github.com/ncruces/go-sqlite3/embed`](https://pkg.go.dev/github.com/ncruces/go-sqlite3/embed)
embeds a build of SQLite into your application.

### Caveats

Because WASM does not support shared memory,
[WAL](https://www.sqlite.org/wal.html) support is [limited](https://www.sqlite.org/wal.html#noshm).

To work around this limitation, SQLite is compiled with
[`SQLITE_DEFAULT_LOCKING_MODE=1`](https://www.sqlite.org/compile.html#default_locking_mode),
making `EXCLUSIVE` the default locking mode.
For non-WAL databases, `NORMAL` locking mode can be activated with
[`PRAGMA locking_mode=NORMAL`](https://www.sqlite.org/pragma.html#pragma_locking_mode).

Because connection pooling is incompatible with `EXCLUSIVE` locking mode,
the `database/sql` driver defaults to `NORMAL` locking mode,
and WAL databases are not supported.

### Roadmap

- [x] build SQLite using `zig cc --target=wasm32-wasi`
- [x] `:memory:` databases
- [x] port [`test_demovfs.c`](https://www.sqlite.org/src/doc/trunk/src/test_demovfs.c) to Go
  - branch [`wasi`](https://github.com/ncruces/go-sqlite3/tree/wasi) uses `test_demovfs.c` directly
- [x] design a nice API, enough for simple use cases
- [x] provide a simple `database/sql` driver
- [x] file locking, compatible with SQLite on macOS/Linux/Windows
- [ ] advanced SQLite features
  - [x] nested transactions
  - [ ] incremental BLOB I/O
  - [ ] online backup
  - [ ] session extension
  - [ ] snapshots
  - [ ] SQL functions
- [ ] custom VFSes
  - [ ] read-only VFS, wrapping an [`io.ReaderAt`](https://pkg.go.dev/io#ReaderAt)
  - [ ] in-memory VFS, wrapping a [`bytes.Buffer`](https://pkg.go.dev/bytes#Buffer)
  - [ ] expose a custom VFS API
  
### Alternatives

- [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite)
- [`crawshaw.io/sqlite`](https://pkg.go.dev/crawshaw.io/sqlite)
- [`github.com/mattn/go-sqlite3`](https://pkg.go.dev/github.com/mattn/go-sqlite3)