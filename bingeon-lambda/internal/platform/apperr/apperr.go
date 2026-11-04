// Package apperr holds the errors that are safe to show in the web app.
// Any other error that reaches the router is logged and reported as
// internal.
package apperr

import (
	"errors"
	"fmt"
	"time"
)

type Code string

const (
	CodeInvalid  Code = "invalid"
	CodeNotFound Code = "not_found"
	// CodeRefused is a rule saying no, usually a limit: Until says when it
	// frees up.
	CodeRefused Code = "refused"
	// CodeConflict is something else changing the same thing at the same
	// moment, more than retrying could get past.
	CodeConflict Code = "conflict"
	// CodeBusy is a service that is rate limiting: try again in a minute.
	CodeBusy Code = "busy"
	// CodeUnavailable is a service bingeon depends on (TMDB, Gemini) failing.
	CodeUnavailable Code = "unavailable"
	// CodeUnauthenticated is a signed-in action called without a user.
	CodeUnauthenticated Code = "unauthenticated"
)

type Error struct {
	Code    Code
	Message string
	// Until is when a refusal frees up, if it will.
	Until *time.Time
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

// Refused is a rule saying no; until is when it frees up (zero: never by
// itself).
func Refused(until time.Time, format string, args ...any) *Error {
	e := &Error{Code: CodeRefused, Message: fmt.Sprintf(format, args...)}
	if !until.IsZero() {
		e.Until = &until
	}
	return e
}

func Invalid(format string, args ...any) *Error {
	return &Error{Code: CodeInvalid, Message: fmt.Sprintf(format, args...)}
}

func NotFound(format string, args ...any) *Error {
	return &Error{Code: CodeNotFound, Message: fmt.Sprintf(format, args...)}
}

func Conflict(format string, args ...any) *Error {
	return &Error{Code: CodeConflict, Message: fmt.Sprintf(format, args...)}
}

func Busy(format string, args ...any) *Error {
	return &Error{Code: CodeBusy, Message: fmt.Sprintf(format, args...)}
}

func Unavailable(format string, args ...any) *Error {
	return &Error{Code: CodeUnavailable, Message: fmt.Sprintf(format, args...)}
}

func Unauthenticated(format string, args ...any) *Error {
	return &Error{Code: CodeUnauthenticated, Message: fmt.Sprintf(format, args...)}
}

// Is reports whether err is an apperr of that code.
func Is(err error, code Code) bool {
	var e *Error
	return errors.As(err, &e) && e.Code == code
}

// Verdict is what an action would say now: allowed, or why not and when
// that frees up. A view carries one per action it offers (its `can`), from
// the same rule the action checks, so the web shows only what the API
// would take.
type Verdict struct {
	OK    bool       `json:"ok"`
	Why   string     `json:"why,omitempty"`
	Until *time.Time `json:"until,omitempty"`
}

// Allowed is the verdict of an action the rules let through.
var Allowed = Verdict{OK: true}

// Err is the refusal an action makes when its verdict says no; nil when
// it says yes.
func (v Verdict) Err() error {
	if v.OK {
		return nil
	}
	return &Error{Code: CodeRefused, Message: v.Why, Until: v.Until}
}

// No is a verdict refusing, until a time (zero: not by itself).
func No(until time.Time, why string) Verdict {
	v := Verdict{Why: why}
	if !until.IsZero() {
		v.Until = &until
	}
	return v
}
