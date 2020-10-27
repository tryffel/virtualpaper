package storage

import (
	"fmt"
	"github.com/lib/pq"
	"strings"
)

type Error struct {
	ErrType string
	ErrMsg  string
	Err     error
}

func newError(errType string) Error {
	return Error{
		ErrType: errType,
		ErrMsg:  errType,
	}
}

func (e Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s - %s: %v", e.ErrType, e.ErrMsg, e.Err)
	}
	return fmt.Sprintf("%s - %s", e.ErrType, e.ErrMsg)
}

var ErrRecordNotFound = newError("not found")
var ErrForbidden = newError("forbidden")
var ErrAlreadyExists = newError("already exists")
var ErrInternalError = newError("internal error")
var ErrInvalid = newError("invalid request")

func getPostgresError(err error) (bool, error) {
	pError, ok := err.(*pq.Error)
	if !ok {
		return false, nil
	}

	// Unique violation
	if pError.Code == "23505" {
		return true, ErrAlreadyExists
	}

	if pError.Code == "23503" {
		return true, ErrRecordNotFound
	}
	return false, err
}

// Catch SQL error, always resulting in internal error
func getSqlError(err error) (bool, error) {
	if strings.Contains(err.Error(), "sql:") {
		if err.Error() == "sql: no rows in result set" {
			return true, ErrRecordNotFound
		}
		return true, fmt.Errorf("%v: %v, ", ErrInternalError, err)
	}
	return false, nil
}

func getDatabaseError(e error) error {
	if e == nil {
		return nil
	}
	p, err := getPostgresError(e)
	if p {
		return err
	}
	sql, err := getSqlError(e)
	if sql {
		return err
	}

	return fmt.Errorf("%v: %v", ErrInternalError, e)
}
