package controller

import (
	"net/http"

	"github.com/dranih/go-crud-api/pkg/record"
)

type Responder interface {
	Error(errorCode int, argument string, w http.ResponseWriter, details string) http.ResponseWriter
	Success(result interface{}, w http.ResponseWriter) http.ResponseWriter
	Exception(err error, w http.ResponseWriter) http.ResponseWriter
	Multi(results []*record.ListDocument, w http.ResponseWriter) http.ResponseWriter
}
