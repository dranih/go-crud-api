package middleware

import (
	"fmt"
	"net/http"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type ApiKeyAuthMiddleware struct {
	GenericMiddleware
}

func NewApiKeyAuth(responder controller.Responder, properties map[string]interface{}) *ApiKeyAuthMiddleware {
	return &ApiKeyAuthMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}}
}

func (akam *ApiKeyAuthMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerName := fmt.Sprint(akam.getProperty("header", "X-API-Key"))
		apiKey := r.Header.Get(headerName)
		ok := true
		if apiKey != "" {
			apiKeys := akam.getArrayProperty("keys", "")
			if val, exists := apiKeys[apiKey]; !exists || !val {
				akam.Responder.Error(record.AUTHENTICATION_FAILED, apiKey, w, "")
				ok = false
			}
		} else {
			if authenticationMode := akam.getProperty("mode", "required"); authenticationMode == "required" {
				realm := fmt.Sprint(akam.getProperty("realm", "Api key required"))
				w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
				akam.Responder.Error(record.AUTHENTICATION_REQUIRED, "", w, realm)
				ok = false
			}
		}
		if ok {
			if apiKey != "" {
				session := utils.GetSession(w, r)
				session.Values["apiKey"] = apiKey
			}
			next.ServeHTTP(w, r)
		}
	})
}
