package middleware

import (
	"net/http"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/record"
)

type AjaxOnlyMiddleware struct {
	GenericMiddleware
}

func NewAjaxOnlyMiddleware(responder controller.Responder, properties map[string]interface{}) *AjaxOnlyMiddleware {
	return &AjaxOnlyMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}}
}

func (aom *AjaxOnlyMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		excludeMethods := aom.getArrayProperty("excludeMethods", "OPTIONS,GET")
		if !excludeMethods[method] {
			headerName := aom.getStringProperty("headerName", "X-Requested-With")
			headerValue := aom.getStringProperty("headerValue", "XMLHttpRequest")
			if headerValue != r.Header.Get(headerName) {
				aom.Responder.Error(record.ONLY_AJAX_REQUESTS_ALLOWED, method, w, "")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
