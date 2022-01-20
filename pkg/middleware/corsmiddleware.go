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
		} else if method == "OPTIONS" {
			allowHeaders := fmt.Sprint(cm.getProperty("allowHeaders", "Content-Type, X-XSRF-TOKEN, X-Authorization"))
			if cm.debug {
				allowHeaders = fmt.Sprintf("%s, %s", allowHeaders, "X-Exception-Name, X-Exception-Message, X-Exception-File")
			}
			if allowHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			}
			if allowedMethods := fmt.Sprint(cm.getProperty("allowMethods", "OPTIONS, GET, PUT, POST, DELETE, PATCH")); allowedMethods != "" {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			}
			if allowCredentials := fmt.Sprint(cm.getProperty("allowMethods", "true")); allowCredentials != "" {
				w.Header().Set("Access-Control-Allow-Credentials", allowCredentials)
			}
			if maxAge := fmt.Sprint(cm.getProperty("maxAge", "1728000")); maxAge != "" {
				w.Header().Set("Access-Control-Allow-Max-Age", maxAge)
			}
			exposeHeaders := fmt.Sprint(cm.getProperty("exposeHeaders", ""))
			if cm.debug {
				exposeHeaders = fmt.Sprintf("%s, %s", exposeHeaders, "X-Exception-Name, X-Exception-Message, X-Exception-File")
			}
			if exposeHeaders != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
			}
			//Returning CORS preflight request
			response := (&controller.ResponseFactory{}).FromStatus(record.OK, w)
			cm.Responder.Success(nil, response)
			return
		} else {
			// else : no origin or origin allowed
			// go on the next middlware in both cases
			if origin != "" {
				if allowCredentials := fmt.Sprint(cm.getProperty("allowMethods", "true")); allowCredentials != "" {
					w.Header().Set("Access-Control-Allow-Credentials", allowCredentials)
				}
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			next.ServeHTTP(w, r)
		}

	})
}

/*
       public function process(ServerRequestInterface $request, RequestHandlerInterface $next): ResponseInterface
       {
           $method = $request->getMethod();
           $origin = count($request->getHeader('Origin')) ? $request->getHeader('Origin')[0] : '';
           $allowedOrigins = $this->getProperty('allowedOrigins', '*');
           if ($origin && !$this->isOriginAllowed($origin, $allowedOrigins)) {
               $response = $this->responder->error(ErrorCode::ORIGIN_FORBIDDEN, $origin);
           } elseif ($method == 'OPTIONS') {
               $response = ResponseFactory::fromStatus(ResponseFactory::OK);
               $allowHeaders = $this->getProperty('allowHeaders', 'Content-Type, X-XSRF-TOKEN, X-Authorization');
               if ($this->debug) {
                   $allowHeaders = implode(', ', array_filter([$allowHeaders, 'X-Exception-Name, X-Exception-Message, X-Exception-File']));
               }
               if ($allowHeaders) {
                   $response = $response->withHeader('Access-Control-Allow-Headers', $allowHeaders);
               }
               $allowMethods = $this->getProperty('allowMethods', 'OPTIONS, GET, PUT, POST, DELETE, PATCH');
               if ($allowMethods) {
                   $response = $response->withHeader('Access-Control-Allow-Methods', $allowMethods);
               }
               $allowCredentials = $this->getProperty('allowCredentials', 'true');
               if ($allowCredentials) {
                   $response = $response->withHeader('Access-Control-Allow-Credentials', $allowCredentials);
               }
               $maxAge = $this->getProperty('maxAge', '1728000');
               if ($maxAge) {
                   $response = $response->withHeader('Access-Control-Max-Age', $maxAge);
               }
               $exposeHeaders = $this->getProperty('exposeHeaders', '');
               if ($this->debug) {
                   $exposeHeaders = implode(', ', array_filter([$exposeHeaders, 'X-Exception-Name, X-Exception-Message, X-Exception-File']));
               }
               if ($exposeHeaders) {
                   $response = $response->withHeader('Access-Control-Expose-Headers', $exposeHeaders);
               }
           } else {
               $response = null;
               try {
                   $response = $next->handle($request);
               } catch (\Throwable $e) {
                   $response = $this->responder->error(ErrorCode::ERROR_NOT_FOUND, $e->getMessage());
                   if ($this->debug) {
                       $response = ResponseUtils::addExceptionHeaders($response, $e);
                   }
               }
           }
           if ($origin) {
               $allowCredentials = $this->getProperty('allowCredentials', 'true');
               if ($allowCredentials) {
                   $response = $response->withHeader('Access-Control-Allow-Credentials', $allowCredentials);
               }
               $response = $response->withHeader('Access-Control-Allow-Origin', $origin);
           }
           return $response;
       }
   }
*/
