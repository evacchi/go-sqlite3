package sqlite3

import (
	"math"
	"time"
)

// Stmt is a prepared statement object.
//
// https://www.sqlite.org/c3ref/stmt.html
type Stmt struct {
	c      *Conn
	handle uint32
	err    error
}

// Close destroys the prepared statement object.
//
// It is safe to close a nil, zero or closed prepared statement.
//
// https://www.sqlite.org/c3ref/finalize.html
func (s *Stmt) Close() error {
	if s == nil || s.handle == 0 {
		return nil
	}

	r, err := s.c.api.finalize.Call(s.c.ctx, uint64(s.handle))
	if err != nil {
		panic(err)
	}

	s.handle = 0
	return s.c.error(r[0])
}

// Reset resets the prepared statement object.
//
// https://www.sqlite.org/c3ref/reset.html
func (s *Stmt) Reset() error {
	r, err := s.c.api.reset.Call(s.c.ctx, uint64(s.handle))
	if err != nil {
		panic(err)
	}
	s.err = nil
	return s.c.error(r[0])
}

// ClearBindings resets all bindings on the prepared statement.
//
// https://www.sqlite.org/c3ref/clear_bindings.html
func (s *Stmt) ClearBindings() error {
	r, err := s.c.api.clearBindings.Call(s.c.ctx, uint64(s.handle))
	if err != nil {
		panic(err)
	}
	return s.c.error(r[0])
}

// Step evaluates the SQL statement.
// If the SQL statement being executed returns any data,
// then true is returned each time a new row of data is ready for processing by the caller.
// The values may be accessed using the Column access functions.
// Step is called again to retrieve the next row of data.
// If an error has occurred, Step returns false;
// call [Stmt.Err] or [Stmt.Reset] to get the error.
//
// https://www.sqlite.org/c3ref/step.html
func (s *Stmt) Step() bool {
	s.c.checkInterrupt()
	r, err := s.c.api.step.Call(s.c.ctx, uint64(s.handle))
	if err != nil {
		panic(err)
	}
	if r[0] == _ROW {
		return true
	}
	if r[0] == _DONE {
		s.err = nil
	} else {
		s.err = s.c.error(r[0])
	}
	return false
}

// Err gets the last error occurred during [Stmt.Step].
// Err returns nil after [Stmt.Reset] is called.
//
// https://www.sqlite.org/c3ref/step.html
func (s *Stmt) Err() error {
	return s.err
}

// Exec is a convenience function that repeatedly calls [Stmt.Step] until it returns false,
// then calls [Stmt.Reset] to reset the statement and get any error that occurred.
func (s *Stmt) Exec() error {
	for s.Step() {
	}
	return s.Reset()
}

// BindCount returns the number of SQL parameters in the prepared statement.
//
// https://www.sqlite.org/c3ref/bind_parameter_count.html
func (s *Stmt) BindCount() int {
	r, err := s.c.api.bindCount.Call(s.c.ctx,
		uint64(s.handle))
	if err != nil {
		panic(err)
	}
	return int(r[0])
}

// BindIndex returns the index of a parameter in the prepared statement
// given its name.
//
// https://www.sqlite.org/c3ref/bind_parameter_index.html
func (s *Stmt) BindIndex(name string) int {
	defer s.c.arena.reset()
	namePtr := s.c.arena.string(name)
	r, err := s.c.api.bindIndex.Call(s.c.ctx,
		uint64(s.handle), uint64(namePtr))
	if err != nil {
		panic(err)
	}
	return int(r[0])
}

// BindName returns the name of a parameter in the prepared statement.
// The leftmost SQL parameter has an index of 1.
//
// https://www.sqlite.org/c3ref/bind_parameter_name.html
func (s *Stmt) BindName(param int) string {
	r, err := s.c.api.bindName.Call(s.c.ctx,
		uint64(s.handle), uint64(param))
	if err != nil {
		panic(err)
	}

	ptr := uint32(r[0])
	if ptr == 0 {
		return ""
	}
	return s.c.mem.readString(ptr, _MAX_STRING)
}

// BindBool binds a bool to the prepared statement.
// The leftmost SQL parameter has an index of 1.
// SQLite does not have a separate boolean storage class.
// Instead, boolean values are stored as integers 0 (false) and 1 (true).
//
// https://www.sqlite.org/c3ref/bind_blob.html
func (s *Stmt) BindBool(param int, value bool) error {
	if value {
		return s.BindInt64(param, 1)
	}
	return s.BindInt64(param, 0)
}

// BindInt binds an int to the prepared statement.
// The leftmost SQL parameter has an index of 1.
//
// https://www.sqlite.org/c3ref/bind_blob.html
func (s *Stmt) BindInt(param int, value int) error {
	return s.BindInt64(param, int64(value))
}

// BindInt64 binds an int64 to the prepared statement.
// The leftmost SQL parameter has an index of 1.
//
// https://www.sqlite.org/c3ref/bind_blob.html
func (s *Stmt) BindInt64(param int, value int64) error {
	r, err := s.c.api.bindInteger.Call(s.c.ctx,
		uint64(s.handle), uint64(param), uint64(value))
	if err != nil {
		panic(err)
	}
	return s.c.error(r[0])
}

// BindFloat binds a float64 to the prepared statement.
// The leftmost SQL parameter has an index of 1.
//
// https://www.sqlite.org/c3ref/bind_blob.html
func (s *Stmt) BindFloat(param int, value float64) error {
	r, err := s.c.api.bindFloat.Call(s.c.ctx,
		uint64(s.handle), uint64(param), math.Float64bits(value))
	if err != nil {
		panic(err)
	}
	return s.c.error(r[0])
}

// BindText binds a string to the prepared statement.
// The leftmost SQL parameter has an index of 1.
//
// https://www.sqlite.org/c3ref/bind_blob.html
func (s *Stmt) BindText(param int, value string) error {
	ptr := s.c.newString(value)
	r, err := s.c.api.bindText.Call(s.c.ctx,
		uint64(s.handle), uint64(param),
		uint64(ptr), uint64(len(value)),
		s.c.api.destructor, _UTF8)
	if err != nil {
		panic(err)
	}
	return s.c.error(r[0])
}

// BindBlob binds a []byte to the prepared statement.
// The leftmost SQL parameter has an index of 1.
// Binding a nil slice is the same as calling [Stmt.BindNull].
//
// https://www.sqlite.org/c3ref/bind_blob.html
func (s *Stmt) BindBlob(param int, value []byte) error {
	ptr := s.c.newBytes(value)
	r, err := s.c.api.bindBlob.Call(s.c.ctx,
		uint64(s.handle), uint64(param),
		uint64(ptr), uint64(len(value)),
		s.c.api.destructor)
	if err != nil {
		panic(err)
	}
	return s.c.error(r[0])
}

// BindZeroBlob binds a zero-filled, length n BLOB to the prepared statement.
// The leftmost SQL parameter has an index of 1.
//
// https://www.sqlite.org/c3ref/bind_blob.html
func (s *Stmt) BindZeroBlob(param int, n int64) error {
	r, err := s.c.api.bindZeroBlob.Call(s.c.ctx,
		uint64(s.handle), uint64(param), uint64(n))
	if err != nil {
		panic(err)
	}
	return s.c.error(r[0])
}

// BindNull binds a NULL to the prepared statement.
// The leftmost SQL parameter has an index of 1.
//
// https://www.sqlite.org/c3ref/bind_blob.html
func (s *Stmt) BindNull(param int) error {
	r, err := s.c.api.bindNull.Call(s.c.ctx,
		uint64(s.handle), uint64(param))
	if err != nil {
		panic(err)
	}
	return s.c.error(r[0])
}

// BindTime binds a [time.Time] to the prepared statement.
// The leftmost SQL parameter has an index of 1.
//
// https://www.sqlite.org/c3ref/bind_blob.html
func (s *Stmt) BindTime(param int, value time.Time, format TimeFormat) error {
	switch v := format.Encode(value).(type) {
	case string:
		s.BindText(param, v)
	case int64:
		s.BindInt64(param, v)
	case float64:
		s.BindFloat(param, v)
	default:
		panic(assertErr())
	}
	return nil
}

// ColumnCount returns the number of columns in a result set.
//
// https://www.sqlite.org/c3ref/column_count.html
func (s *Stmt) ColumnCount() int {
	r, err := s.c.api.columnCount.Call(s.c.ctx,
		uint64(s.handle))
	if err != nil {
		panic(err)
	}
	return int(r[0])
}

// ColumnName returns the name of the result column.
// The leftmost column of the result set has the index 0.
//
// https://www.sqlite.org/c3ref/column_name.html
func (s *Stmt) ColumnName(col int) string {
	r, err := s.c.api.columnName.Call(s.c.ctx,
		uint64(s.handle), uint64(col))
	if err != nil {
		panic(err)
	}

	ptr := uint32(r[0])
	if ptr == 0 {
		panic(oomErr)
	}
	return s.c.mem.readString(ptr, _MAX_STRING)
}

// ColumnType returns the initial [Datatype] of the result column.
// The leftmost column of the result set has the index 0.
//
// https://www.sqlite.org/c3ref/column_blob.html
func (s *Stmt) ColumnType(col int) Datatype {
	r, err := s.c.api.columnType.Call(s.c.ctx,
		uint64(s.handle), uint64(col))
	if err != nil {
		panic(err)
	}
	return Datatype(r[0])
}

// ColumnBool returns the value of the result column as a bool.
// The leftmost column of the result set has the index 0.
// SQLite does not have a separate boolean storage class.
// Instead, boolean values are retrieved as integers,
// with 0 converted to false and any other value to true.
//
// https://www.sqlite.org/c3ref/column_blob.html
func (s *Stmt) ColumnBool(col int) bool {
	if i := s.ColumnInt64(col); i != 0 {
		return true
	}
	return false
}

// ColumnInt returns the value of the result column as an int.
// The leftmost column of the result set has the index 0.
//
// https://www.sqlite.org/c3ref/column_blob.html
func (s *Stmt) ColumnInt(col int) int {
	return int(s.ColumnInt64(col))
}

// ColumnInt64 returns the value of the result column as an int64.
// The leftmost column of the result set has the index 0.
//
// https://www.sqlite.org/c3ref/column_blob.html
func (s *Stmt) ColumnInt64(col int) int64 {
	r, err := s.c.api.columnInteger.Call(s.c.ctx,
		uint64(s.handle), uint64(col))
	if err != nil {
		panic(err)
	}
	return int64(r[0])
}

// ColumnFloat returns the value of the result column as a float64.
// The leftmost column of the result set has the index 0.
//
// https://www.sqlite.org/c3ref/column_blob.html
func (s *Stmt) ColumnFloat(col int) float64 {
	r, err := s.c.api.columnFloat.Call(s.c.ctx,
		uint64(s.handle), uint64(col))
	if err != nil {
		panic(err)
	}
	return math.Float64frombits(r[0])
}

// ColumnTime returns the value of the result column as a [time.Time].
// The leftmost column of the result set has the index 0.
//
// https://www.sqlite.org/c3ref/column_blob.html
func (s *Stmt) ColumnTime(col int, format TimeFormat) time.Time {
	var v any
	switch s.ColumnType(col) {
	case INTEGER:
		v = s.ColumnInt64(col)
	case FLOAT:
		v = s.ColumnFloat(col)
	case TEXT, BLOB:
		v = s.ColumnText(col)
	case NULL:
		return time.Time{}
	default:
		panic(assertErr())
	}
	t, err := format.Decode(v)
	if err != nil {
		s.err = err
	}
	return t
}

// ColumnText returns the value of the result column as a string.
// The leftmost column of the result set has the index 0.
//
// https://www.sqlite.org/c3ref/column_blob.html
func (s *Stmt) ColumnText(col int) string {
	r, err := s.c.api.columnText.Call(s.c.ctx,
		uint64(s.handle), uint64(col))
	if err != nil {
		panic(err)
	}

	ptr := uint32(r[0])
	if ptr == 0 {
		r, err = s.c.api.errcode.Call(s.c.ctx, uint64(s.handle))
		if err != nil {
			panic(err)
		}
		s.err = s.c.error(r[0])
		return ""
	}

	r, err = s.c.api.columnBytes.Call(s.c.ctx,
		uint64(s.handle), uint64(col))
	if err != nil {
		panic(err)
	}

	mem := s.c.mem.view(ptr, uint32(r[0]))
	return string(mem)
}

// ColumnBlob appends to buf and returns
// the value of the result column as a []byte.
// The leftmost column of the result set has the index 0.
//
// https://www.sqlite.org/c3ref/column_blob.html
func (s *Stmt) ColumnBlob(col int, buf []byte) []byte {
	r, err := s.c.api.columnBlob.Call(s.c.ctx,
		uint64(s.handle), uint64(col))
	if err != nil {
		panic(err)
	}

	ptr := uint32(r[0])
	if ptr == 0 {
		r, err = s.c.api.errcode.Call(s.c.ctx, uint64(s.handle))
		if err != nil {
			panic(err)
		}
		s.err = s.c.error(r[0])
		return buf[0:0]
	}

	r, err = s.c.api.columnBytes.Call(s.c.ctx,
		uint64(s.handle), uint64(col))
	if err != nil {
		panic(err)
	}

	mem := s.c.mem.view(ptr, uint32(r[0]))
	return append(buf[0:0], mem...)
}

// Return true if stmt is an empty SQL statement.
// This is used as an optimization.
// It's OK to always return false here.
func emptyStatement(stmt string) bool {
	for _, b := range []byte(stmt) {
		switch b {
		case ' ', '\n', '\r', '\t', '\v', '\f':
		case ';':
		default:
			return false
		}
	}
	return true
}
