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
		return fmt.Sprintf("%s - %s: %s", e.ErrType, e.ErrMsg, e.Err.Error())
	}
	return fmt.Sprintf("%s - %s", e.ErrType, e.ErrMsg)
}

var ErrRecordNotFound = newError("not found")
var ErrForbidden = newError("forbidden")
var ErrAlreadyExists = newError("already exists")
var ErrInternalError = newError("internal error")
var ErrInvalid = newError("invalid request")

func getPostgresError(err error, resource string, action string) (bool, error) {
	pError, ok := err.(*pq.Error)
	if !ok {
		return false, nil
	}

	// Unique violation
	if pError.Code == "23505" {
		e := ErrAlreadyExists
		e.ErrMsg = resource + " already exists"
		return true, e
	}

	if pError.Code == "23503" {
		e := ErrRecordNotFound
		e.ErrMsg = resource + " not found"
		return true, e
	}

	e := ErrInternalError
	e.ErrMsg = pError.Message
	e.Err = err
	return true, e
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

func getDatabaseError(e error, resource string, action string) error {
	if e == nil {
		return nil
	}
	p, err := getPostgresError(e, resource, action)
	if p {
		return err
	}
	sql, err := getSqlError(e)
	if sql {
		return err
	}
	Err := ErrInternalError
	Err.Err = err
	Err.ErrMsg = fmt.Sprintf("%s - %s", resource, action)
	return Err
}
