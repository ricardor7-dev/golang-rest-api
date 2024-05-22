package errs

import (
	"errors"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"

	pkgerrors "github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const EMINIOBUCKETEXITS	=	"bucket_exists"
type Error struct {
	// Op is the operation being performed, usually the name of the method
	// being invoked.
	Op Op
	// Code is a human-readable, short representation of the error
	Code Code		
	//Exteral is a json format error from External services
	External json.RawMessage
	//standard error, to be use in the message
	Err error
}

type Op string

// Unwrap method allows for unwrapping errors using errors.As
func (e Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func E(args ...interface{}) error {
	type stackTracer interface {
		StackTrace() pkgerrors.StackTrace
	}

	if len(args) == 0 {
		panic("call to errors.E with no arguments")
	}
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.Op = arg
		case Code:
			e.Code = arg
			if zerolog.ErrorStackMarshaler != nil {
				e.Err = pkgerrors.New(e.Code.Code)
			} else {
				e.Err = Str(e.Code.Code)
			}
		case string:
			if zerolog.ErrorStackMarshaler != nil {
				e.Err = pkgerrors.New(arg)
			} else {
				e.Err = Str(arg)
			}
		case json.RawMessage:
			e.External = arg
		case *Error:
			// Make a copy
			errorCopy := *arg
			e.Err = &errorCopy
		case error:
			if zerolog.ErrorStackMarshaler != nil {
				// if the error implements stackTracer, then it is
				// a pkg/errors error type and does not need to have
				// the stack added
				_, ok := arg.(stackTracer)
				if ok {
					e.Err = arg
				} else {
					e.Err = pkgerrors.New(arg.Error())
				}
			} else {
				e.Err = arg
			}
		default:		//remover??
			_, file, line, _ := runtime.Caller(1)
			return fmt.Errorf("errors.E: bad call from %s:%d: %v, unknown type %T, value %v in error call", file, line, args, arg, arg)
		}
	}
	
	prev, ok := e.Err.(*Error)
	if !ok {
		return e
	}

	if prev.Code == e.Code {
		prev.Code = Code{}
	}
	// If this error has Code == "", pull up the inner one.
	if e.Code == (Code{}) {
		e.Code = prev.Code
		prev.Code = Code{}
	}

	return e
}


func (e *Error) isZero() bool {
	return e.Code.Code == "" && e.Code.Info == "" && e.Err == nil
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}
// Str returns an error that formats as the given text. It is intended to
// be used as the error-typed argument to the E function.
func Str(text string) error {
	return &errorString{text}
}

func (e *errorString) Error() string {
	return e.s
}

//OpStack returns the op stack information for an error
func OpStack(err error) []string {
	type o struct {
		Op    string
		Order int
	}

	e := err
	i := 0
	var os []o

	// loop through all wrapped errors and add to struct
	// order will be from top to bottom of stack
	for errors.Unwrap(e) != nil {
		var errsError *Error
		if errors.As(e, &errsError) {
			if errsError.Op != "" {
				op := o{Op: string(errsError.Op), Order: i}
				os = append(os, op)
			}
		}
		e = errors.Unwrap(e)
		i++
	}

	// reverse the order of the stack (bottom to top)
	sort.Slice(os, func(i, j int) bool { return os[i].Order > os[j].Order })

	// pull out just the stack info, now in reversed order
	var ops []string
	for _, op := range os {
		ops = append(ops, op.Op)
	}

	return ops
}

// TopError recursively unwraps all errors and retrieves the topmost error
func TopError(err error) error {
	currentErr := err
	for errors.Unwrap(currentErr) != nil {
		currentErr = errors.Unwrap(currentErr)
	}

	return currentErr
}

func Is(err, targert error) bool{
	return errors.Is(err, targert)
}