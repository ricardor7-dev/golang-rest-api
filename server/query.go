package server

import (
	"errors"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"

	"golang-rest-api/errs"
)

// created once, both are safe for concurrent use
var queryDecoder = newQueryDecoder()
var queryValidator = newQueryValidator()

func newQueryDecoder() *schema.Decoder {
	d := schema.NewDecoder()
	d.IgnoreUnknownKeys(true)
	return d
}

// validation errors report the query parameter name (schema tag) instead of the Go field name
func newQueryValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		return strings.Split(f.Tag.Get("schema"), ",")[0]
	})
	return v
}

// queryParamCodes maps each query parameter to the error code sent when it's invalid
var queryParamCodes = map[string]errs.Code{
	"tag":          errs.AB_ERR_002,
	"page":         errs.AB_ERR_003,
	"itemsPerPage": errs.AB_ERR_004,
	"sortBy":       errs.AB_ERR_017,
	"orderBy":      errs.AB_ERR_018,
	"date":         errs.AB_ERR_019,
}

// decodeQuery fills dst from the query string and validates it
func decodeQuery(values url.Values, dst interface{}) error {
	if err := queryDecoder.Decode(dst, values); err != nil {
		var multi schema.MultiError
		if errors.As(err, &multi) {
			for param, e := range multi {
				return errs.E(queryParamCode(param), e)
			}
		}
		return errs.E(errs.AB_ERR_020, err)
	}

	if err := queryValidator.Struct(dst); err != nil {
		var fieldErrs validator.ValidationErrors
		if errors.As(err, &fieldErrs) && len(fieldErrs) > 0 {
			return errs.E(queryParamCode(fieldErrs[0].Field()), err)
		}
		return errs.E(errs.AB_ERR_020, err)
	}

	return nil
}

func queryParamCode(param string) errs.Code {
	if code, ok := queryParamCodes[param]; ok {
		return code
	}
	return errs.AB_ERR_020
}
