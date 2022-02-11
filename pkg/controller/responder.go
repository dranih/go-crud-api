package controller

import (
	"net/http"
)

type Responder interface {
	Error(errorCode int, argument string, w http.ResponseWriter, r *http.Request, details interface{}) http.ResponseWriter
	Success(result interface{}, w http.ResponseWriter, r *http.Request) http.ResponseWriter
	Exception(err error, w http.ResponseWriter, r *http.Request) http.ResponseWriter
	Multi(results *[]interface{}, errs []error, w http.ResponseWriter, r *http.Request) http.ResponseWriter
	SetAfterHandler(handler string) error
}
