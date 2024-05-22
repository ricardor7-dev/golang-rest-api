package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

// ErrResponse is used as the Response Body
type ErrResponse struct {
	Error 	Status 		`json:"status"`
	Result	interface{}	`json:"result"`
}

// ServiceError has fields for Service errors. All fields with no data will
// be omitted
type Status struct {
	Code    		string `json:"code,omitempty"`
	Message 		string `json:"message,omitempty"`
	Description   	string `json:"description,omitempty"`
}

func HTTPErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, err error) {
	if err == nil {
		nilErrorResponse(w, lgr)
		return
	}

	var e *Error
	if errors.As(err, &e) {
		switch e.Code.Code {

		default:
			typicalErrorResponse(w, lgr, e)
			return
		}
	}

	unknownErrorResponse(w, lgr, err)
}

// httpErrorStatusCode maps an error Kind to an HTTP Status Code
func httpErrorStatusCode(k Code) int {
	switch k {
	case AB_ERR_002, AB_ERR_003, AB_ERR_004, AB_ERR_005,
	AB_ERR_006, AB_ERR_011, AB_ERR_015, AB_ERR_016:
		return http.StatusBadRequest
	case AB_ERR_014:
		return http.StatusConflict
	case AB_ERR_001:
		return http.StatusNotFound
	case AB_ERR_009:
		return http.StatusRequestEntityTooLarge
	case AB_ERR_010:
		return http.StatusUnprocessableEntity
	case AB_ERR_007:
		return http.StatusServiceUnavailable
	case AB_ERR_008, AB_ERR_012:
		return http.StatusBadGateway		//TODO: IS this  the correct code?
	case AB_ERR_013:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// typicalErrorResponse replies to the request with the specified error
// message and HTTP code. It does not otherwise end the request; the
// caller should ensure no further writes are done to w.
//
// Taken from standard library and modified.
// https://golang.org/pkg/net/http/#Error
func typicalErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, e *Error) {

	httpStatusCode := httpErrorStatusCode(e.Code)

	// We can retrieve the status here and write out a specific
	// HTTP status code. If the error is empty, just send the HTTP
	// Status Code as response. Error should not be empty, but it's
	// theoretically possible, so this is just in case...
	if e.isZero() {
		lgr.Error().Stack().Int("http_statuscode", httpStatusCode).Msg("empty error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// typical errors
	const errMsg = "error response sent to client"
	if zerolog.ErrorStackMarshaler != nil {
		err := TopError(e)

		// log the error with stacktrace from "github.com/pkg/errors"
		// do not bother to log with op stack
		lgr.Error().Stack().Err(err).
			Int("http_statuscode", httpStatusCode).
			Str("code", string(e.Code.Code)).
			Str("message", string(e.Code.Info)).
			Msg(errMsg)
	} else {
		ops := OpStack(e)
		if len(ops) > 0 {
			j, _ := json.Marshal(ops)
			// log the error with the op stack
			lgr.Error().RawJSON("stack", j).Err(e.Err).
				Int("http_statuscode", httpStatusCode).
				Str("code", string(e.Code.Code)).
				Str("message", string(e.Code.Info)).
				Msg(errMsg)
		} else {
			// no op stack present, log the error without that field
			lgr.Error().Err(e.Err).
				Int("http_statuscode", httpStatusCode).
				Str("code", string(e.Code.Code)).
				Str("message", string(e.Code.Info)).
				Msg(errMsg)
		}
	}

	// get ErrResponse
	er := newErrResponse(e)
	response := ErrResponse{
		Error: er,
	}
	// Marshal errResponse struct to JSON for the response body
	errJSON, _ := json.Marshal(response)
	ej := string(errJSON)

	// Write Content-Type headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Write HTTP Statuscode
	w.WriteHeader(httpStatusCode)

	// Write response body (json)
	fmt.Fprintln(w, ej)
}

func newErrResponse(err *Error) Status {
	const msg string = "internal server error - please contact support"

	return Status{ 
		Code: err.Code.Code,
		Message: err.Code.Info,
		Description: err.Error(),
	}
}

// nilErrorResponse responds with http status code 500 (Internal Server Error)
// and an empty response body. nil error should never be sent, but in case it is...
func nilErrorResponse(w http.ResponseWriter, lgr zerolog.Logger) {
	lgr.Error().Stack().
		Int("HTTP Error StatusCode", http.StatusInternalServerError).
		Msg("nil error - no response body sent")

	w.WriteHeader(http.StatusInternalServerError)
}

// unknownErrorResponse responds with http status code 500 (Internal Server Error)
// and a json response body with unanticipated_error kind
func unknownErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, err error) {
	er := ErrResponse{
		Error: Status{
			Code:    ERR_UNANTICIPATED.Code,
			Message: ERR_UNANTICIPATED.Info,
		},
	}

	lgr.Error().Err(err).Msg("Unknown Error")

	// Marshal errResponse struct to JSON for the response body
	errJSON, _ := json.Marshal(er)
	ej := string(errJSON)

	// Write Content-Type headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Write HTTP Statuscode
	w.WriteHeader(http.StatusInternalServerError)

	// Write response body (json)
	fmt.Fprintln(w, ej)
}