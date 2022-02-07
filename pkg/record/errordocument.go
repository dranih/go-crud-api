package record

import (
	"encoding/json"
	"strings"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
)

type ErrorDocument struct {
	errorCode *ErrorCode
	argument  string
	details   interface{}
}

func NewErrorDocument(errorCode *ErrorCode, argument string, details interface{}) *ErrorDocument {
	return &ErrorDocument{errorCode, argument, details}
}

func (ed *ErrorDocument) GetStatus() int {
	return ed.errorCode.GetStatus()
}

func (ed *ErrorDocument) GetCode() int {
	return ed.errorCode.GetCode()
}

func (ed *ErrorDocument) GetMessage() string {
	return ed.errorCode.GetMessage(ed.argument)
}

func (ed *ErrorDocument) Serialize() map[string]interface{} {
	if ed.details == "" {
		return map[string]interface{}{"code": ed.GetCode(),
			"message": ed.GetMessage(),
		}
	}
	return map[string]interface{}{"code": ed.GetCode(),
		"message": ed.GetMessage(),
		"details": ed.details}
}

// json marshaling for struct ErrorDocument
func (ed *ErrorDocument) MarshalJSON() ([]byte, error) {
	return json.Marshal(ed.Serialize())
}

func NewErrorDocumentFromError(err error, debug bool) *ErrorDocument {
	switch err.(type) {
	case sqlite3.Error, *pq.Error, *mysql.MySQLError, mssql.Error:
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return NewErrorDocument(NewErrorCode(DUPLICATE_KEY_EXCEPTION), "", "")
		} else if strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
			return NewErrorDocument(NewErrorCode(DUPLICATE_KEY_EXCEPTION), "", "")
		} else if strings.Contains(strings.ToLower(err.Error()), "default value") {
			return NewErrorDocument(NewErrorCode(DATA_INTEGRITY_VIOLATION), "", "")
		} else if strings.Contains(strings.ToLower(err.Error()), "allow nulls") {
			return NewErrorDocument(NewErrorCode(DATA_INTEGRITY_VIOLATION), "", "")
		} else if strings.Contains(strings.ToLower(err.Error()), "constraint") {
			return NewErrorDocument(NewErrorCode(DATA_INTEGRITY_VIOLATION), "", "")
		} else {
			message := "SQL exception occurred (enable debug mode)"
			if debug {
				message = err.Error()
			}
			return NewErrorDocument(NewErrorCode(ERROR_NOT_FOUND), message, "")
		}
	}
	return NewErrorDocument(NewErrorCode(ERROR_NOT_FOUND), err.Error(), "")
}
