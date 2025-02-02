package sqlite3

import (
	"runtime"
	"strconv"
	"strings"
)

// Error wraps an SQLite Error Code.
//
// https://www.sqlite.org/c3ref/errcode.html
type Error struct {
	code uint64
	str  string
	msg  string
	sql  string
}

// Code returns the primary error code for this error.
//
// https://www.sqlite.org/rescode.html
func (e *Error) Code() ErrorCode {
	return ErrorCode(e.code)
}

// ExtendedCode returns the extended error code for this error.
//
// https://www.sqlite.org/rescode.html
func (e *Error) ExtendedCode() ExtendedErrorCode {
	return ExtendedErrorCode(e.code)
}

// Error implements the error interface.
func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString("sqlite3: ")

	if e.str != "" {
		b.WriteString(e.str)
	} else {
		b.WriteString(strconv.Itoa(int(e.code)))
	}

	if e.msg != "" {
		b.WriteByte(':')
		b.WriteByte(' ')
		b.WriteString(e.msg)
	}

	return b.String()
}

// Temporary returns true for [BUSY] errors.
func (e *Error) Temporary() bool {
	return e.Code() == BUSY
}

// SQL returns the SQL starting at the token that triggered a syntax error.
func (e *Error) SQL() string {
	return e.sql
}

type errorString string

func (e errorString) Error() string { return string(e) }

const (
	binaryErr   = errorString("sqlite3: no SQLite binary embed/set/loaded")
	nilErr      = errorString("sqlite3: invalid memory address or null pointer dereference")
	oomErr      = errorString("sqlite3: out of memory")
	rangeErr    = errorString("sqlite3: index out of range")
	noNulErr    = errorString("sqlite3: missing NUL terminator")
	noGlobalErr = errorString("sqlite3: could not find global: ")
	noFuncErr   = errorString("sqlite3: could not find function: ")
	timeErr     = errorString("sqlite3: invalid time value")
	notImplErr  = errorString("sqlite3: not implemented")
)

func assertErr() errorString {
	msg := "sqlite3: assertion failed"
	if _, file, line, ok := runtime.Caller(1); ok {
		msg += " (" + file + ":" + strconv.Itoa(line) + ")"
	}
	return errorString(msg)
}
