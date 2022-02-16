package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/record"
)

type CorsMiddleware struct {
	GenericMiddleware
	debug bool
}

func NewCorsMiddleware(responder controller.Responder, properties map[string]interface{}, debug bool) *CorsMiddleware {
	return &CorsMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, debug: debug}
}

func (cm *CorsMiddleware) isOriginAllowed(origin, allowedOrigins string) bool {
	for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
		hostname := regexp.QuoteMeta(strings.ToLower(strings.TrimSpace(allowedOrigin)))
		if r, err := regexp.Compile(fmt.Sprintf("^%s$", strings.Replace(hostname, `\*`, `.*`, -1))); err == nil {
			if r.Match([]byte(origin)) {
				return true
			}
		}
	}
	return false
}

func (cm *CorsMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		origin := r.Header.Get("Origin")
		allowedOrigins := fmt.Sprint(cm.getProperty("allowedOrigins", `*`))
		// if origin header and not allowed => Forbidden
		if origin != "" && !cm.isOriginAllowed(origin, allowedOrigins) {
			cm.Responder.Error(record.ORIGIN_FORBIDDEN, origin, w, "")
			return
		} else if strings.ToUpper(method) == "OPTIONS" {
			allowHeaders := fmt.Sprint(cm.getProperty("allowHeaders", "Content-Type, X-XSRF-TOKEN, X-Authorization"))
			if cm.debug {
				allowHeaders = fmt.Sprintf("%s, %s", allowHeaders, "X-Exception-Name, X-Exception-Message, X-Exception-File")
			}
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			w.Header().Set("Access-Control-Allow-Methods", fmt.Sprint(cm.getProperty("allowMethods", "OPTIONS, GET, PUT, POST, DELETE, PATCH")))
			w.Header().Set("Access-Control-Allow-Credentials", fmt.Sprint(cm.getProperty("allowMethods", "true")))
			w.Header().Set("Access-Control-Max-Age", fmt.Sprint(cm.getProperty("maxAge", "1728000")))
			exposeHeaders := fmt.Sprint(cm.getProperty("exposeHeaders", ""))
			if cm.debug {
				exposeHeaders = fmt.Sprintf("%s, %s", exposeHeaders, "X-Exception-Name, X-Exception-Message, X-Exception-File")
			}
			w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Credentials", fmt.Sprint(cm.getProperty("allowMethods", "true")))
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			//Returning CORS preflight request
			(&controller.ResponseFactory{}).FromStatus(record.OK, w)
			return
		} else {
			// else : no origin or origin allowed
			// go on the next middlware in both cases
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Credentials", fmt.Sprint(cm.getProperty("allowMethods", "true")))
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			next.ServeHTTP(w, r)
		}
	})
}
