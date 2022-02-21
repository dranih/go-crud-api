package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dranih/go-crud-api/pkg/controller"
)

type SslRedirectMiddleware struct {
	GenericMiddleware
	httpsPort int
}

func NewSslRedirectMiddleware(responder controller.Responder, properties map[string]interface{}, httpsPort int) *SslRedirectMiddleware {
	return &SslRedirectMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, httpsPort: httpsPort}
}

func (srm *SslRedirectMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil || r.URL.Scheme == "http" {
			host := strings.Split(r.Host, ":")[0]
			http.Redirect(w, r, fmt.Sprintf("https://%s:%d%s", host, srm.httpsPort, r.URL.Path), http.StatusMovedPermanently)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
