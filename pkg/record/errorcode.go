package record

import "fmt"

type ErrorCode struct {
	code    int
	message string
	status  int
}

const OK = 200
const MOVED_PERMANENTLY = 301
const FOUND = 302
const UNAUTHORIZED = 401
const FORBIDDEN = 403
const NOT_FOUND = 404
const METHOD_NOT_ALLOWED = 405
const CONFLICT = 409
const UNPROCESSABLE_ENTITY = 422
const FAILED_DEPENDENCY = 424
const INTERNAL_SERVER_ERROR = 500

const ERROR_NOT_FOUND = 9999
const ROUTE_NOT_FOUND = 1000
const TABLE_NOT_FOUND = 1001
const ARGUMENT_COUNT_MISMATCH = 1002
const RECORD_NOT_FOUND = 1003
const ORIGIN_FORBIDDEN = 1004
const COLUMN_NOT_FOUND = 1005
const TABLE_ALREADY_EXISTS = 1006
const COLUMN_ALREADY_EXISTS = 1007
const HTTP_MESSAGE_NOT_READABLE = 1008
const DUPLICATE_KEY_EXCEPTION = 1009
const DATA_INTEGRITY_VIOLATION = 1010
const AUTHENTICATION_REQUIRED = 1011
const AUTHENTICATION_FAILED = 1012
const INPUT_VALIDATION_FAILED = 1013
const OPERATION_FORBIDDEN = 1014
const OPERATION_NOT_SUPPORTED = 1015
const TEMPORARY_OR_PERMANENTLY_BLOCKED = 1016
const BAD_OR_MISSING_XSRF_TOKEN = 1017
const ONLY_AJAX_REQUESTS_ALLOWED = 1018
const PAGINATION_FORBIDDEN = 1019
const USER_ALREADY_EXIST = 1020
const PASSWORD_TOO_SHORT = 1021

func NewErrorCode(code int) *ErrorCode {
	values := map[int][]interface{}{
		0000: {"Success", OK},
		1000: {"Route '%s' not found", NOT_FOUND},
		1001: {"Table '%s' not found", NOT_FOUND},
		1002: {"Argument count mismatch in '%s'", UNPROCESSABLE_ENTITY},
		1003: {"Record '%s' not found", NOT_FOUND},
		1004: {"Origin '%s' is forbidden", FORBIDDEN},
		1005: {"Column '%s' not found", NOT_FOUND},
		1006: {"Table '%s' already exists", CONFLICT},
		1007: {"Column '%s' already exists", CONFLICT},
		1008: {"Cannot read HTTP message", UNPROCESSABLE_ENTITY},
		1009: {"Duplicate key exception", CONFLICT},
		1010: {"Data integrity violation", CONFLICT},
		1011: {"Authentication required", UNAUTHORIZED},
		1012: {"Authentication failed for '%s'", FORBIDDEN},
		1013: {"Input validation failed for '%s'", UNPROCESSABLE_ENTITY},
		1014: {"Operation forbidden", FORBIDDEN},
		1015: {"Operation '%s' not supported", METHOD_NOT_ALLOWED},
		1016: {"Temporary or permanently blocked", FORBIDDEN},
		1017: {"Bad or missing XSRF token", FORBIDDEN},
		1018: {"Only AJAX requests allowed for '%s'", FORBIDDEN},
		1019: {"Pagination forbidden", FORBIDDEN},
		1020: {"User '%s' already exists", CONFLICT},
		1021: {"Password too short (<%d characters)", UNPROCESSABLE_ENTITY},
		9999: {"%s", INTERNAL_SERVER_ERROR},
	}
	if _, b := values[code]; !b {
		code = 9999
	}
	return &ErrorCode{code, values[code][0].(string), values[code][1].(int)}
}

func (ec *ErrorCode) GetCode() int {
	return ec.code
}

func (ec *ErrorCode) GetMessage(argument string) string {
	return fmt.Sprintf(ec.message, argument)
}

func (ec *ErrorCode) GetStatus() int {
	return ec.status
}
