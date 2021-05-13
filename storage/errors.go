package storage

import (
	"fmt"
	"github.com/lib/pq"
	"strings"
	"tryffel.net/go/virtualpaper/errors"
)

func getPostgresError(err error, resource Resource, action string) (bool, error) {
	pError, ok := err.(*pq.Error)
	if !ok {
		return false, nil
	}

	// Unique violation
	if pError.Code == "23505" {
		e := errors.ErrAlreadyExists
		e.ErrMsg = resource.Name() + " already exists"
		return true, e
	}

	if pError.Code == "23503" {
		e := errors.ErrRecordNotFound
		e.ErrMsg = resource.Name() + " not found"
		return true, e
	}

	e := errors.ErrInternalError
	e.ErrMsg = pError.Message
	e.Err = err
	return true, e
}

// Catch SQL error, always resulting in internal error
func getSqlError(err error, resource Resource) (bool, error) {
	if strings.Contains(err.Error(), "sql:") {
		if err.Error() == "sql: no rows in result set" {
			e := errors.ErrRecordNotFound
			e.ErrMsg = resource.Name() + " not found"
			return true, e
		}
		return true, fmt.Errorf("%v: %v, ", errors.ErrInternalError, err)
	}
	return false, nil
}

func getDatabaseError(e error, resource Resource, action string) error {
	if e == nil {
		return nil
	}
	p, err := getPostgresError(e, resource, action)
	if p {
		return err
	}
	sql, err := getSqlError(e, resource)
	if sql {
		return err
	}
	Err := errors.ErrInternalError
	Err.Err = err
	Err.ErrMsg = fmt.Sprintf("%s - %s", resource, action)
	return Err
}

func getDatabaseErrorIgnoreEmpty(e error, resource Resource, action string) error {
	if e == nil {
		return nil
	}
	p, err := getPostgresError(e, resource, action)
	if p {
		if e, ok := err.(errors.Error); ok {
			if e.Is(errors.ErrRecordNotFound) {
				return nil
			}
		}
		return err
	}
	sql, err := getSqlError(e, resource)
	if sql {
		return err
	}
	Err := errors.ErrInternalError
	Err.Err = err
	Err.ErrMsg = fmt.Sprintf("%s - %s", resource, action)
	return Err
}
