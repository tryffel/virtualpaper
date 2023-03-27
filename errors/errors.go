/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2020  Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

/*
	Package errors contains application specific error specification

and resolution.
*/
package errors

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
)

var New = errors.New

var Is = errors.Is

var As = errors.As

var Unwrap = errors.Unwrap

type Error struct {
	ErrType string
	ErrMsg  string
	Err     error
	Stack   []byte
}

func newError(errType string) Error {
	return Error{
		ErrType: errType,
		ErrMsg:  errType,
	}
}

func (e *Error) SetStack() {
	e.Stack = []byte(getStack(8))
}

// getStack returns stack trace.
// Set skip>2 to remove not-interesting lines, such as (errors.go/getStack()).
func getStack(skip int) string {
	rawStack := string(debug.Stack())

	// remove top two functions from stack, that is, debug.Stack, task.recoverPanic && Panic
	lines := strings.Split(rawStack, "\n")
	// goroutine num
	stack := lines[0]

	prints := lines[skip:]
	for _, v := range prints {
		stack = stack + "\n" + v
	}
	return stack
}

func (e Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s - %s: %s", e.ErrType, e.ErrMsg, e.Err.Error())
	}
	return fmt.Sprintf("%s - %s", e.ErrType, e.ErrMsg)
}

func (e Error) Is(target error) bool {
	eTarget, ok := target.(Error)
	if !ok {
		return false
	}

	return eTarget.ErrType == e.ErrType
}

var ErrRecordNotFound = newError("not found")
var ErrForbidden = newError("forbidden")
var ErrUnauthorized = newError("unauthorized")
var ErrAlreadyExists = newError("already exists")
var ErrInternalError = newError("internal error")
var ErrInvalid = newError("invalid request")
