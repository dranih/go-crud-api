package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/record"
)

type XsrfMiddleware struct {
	GenericMiddleware
}

func NewXsrfMiddleware(responder controller.Responder, properties map[string]interface{}) *XsrfMiddleware {
	return &XsrfMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}}
}

func (xm *XsrfMiddleware) generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (xm *XsrfMiddleware) getToken(w http.ResponseWriter, r *http.Request) string {
	cookieName := xm.getStringProperty("cookieName", "XSRF-TOKEN")
	cookie, err := r.Cookie(cookieName)
	if err == nil && cookie != nil {
		return cookie.Value
	} else {
		if tokenB, err := xm.generateRandomBytes(8); err == nil {
			token := hex.EncodeToString(tokenB)
			expiration := time.Now().Add(24 * time.Hour)
			cookie := http.Cookie{Name: cookieName, Value: token, Expires: expiration}
			http.SetCookie(w, &cookie)
			return token
		}
		return ""
	}
}

func (xm *XsrfMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := xm.getToken(w, r)
		if token == "" {
			xm.Responder.Error(record.INTERNAL_SERVER_ERROR, "", w, "")
			return
		}
		method := r.Method
		excludeMethods := xm.getArrayProperty("excludeMethods", "OPTIONS,GET")
		if !excludeMethods[method] {
			headerName := xm.getStringProperty("headerName", "X-XSRF-TOKEN")
			if token != r.Header.Get(headerName) {
				xm.Responder.Error(record.BAD_OR_MISSING_XSRF_TOKEN, "", w, "")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
