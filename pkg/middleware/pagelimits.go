package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type PageLimitsMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
}

func NewPageLimitsMiddleware(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService) *PageLimitsMiddleware {
	return &PageLimitsMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection}
}

func (pm *PageLimitsMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operation := utils.GetOperation(r)
		if operation == "list" {
			params := utils.GetRequestParams(r)
			maxPage := pm.getIntProperty("pages", 100)
			if v, ok := params["page"]; ok && len(v) > 0 && maxPage > 0 {
				var page int
				var err error
				if !strings.Contains(v[0], ",") {
					page, err = strconv.Atoi(v[0])
				} else {
					page, err = strconv.Atoi(strings.SplitN(v[0], ",", 2)[0])
				}
				if err == nil && page > maxPage {
					pm.Responder.Error(record.PAGINATION_FORBIDDEN, "", w, nil)
					return
				}
			}

			maxSize := pm.getIntProperty("records", 1000)
			if v, ok := params["size"]; (!ok || len(v) == 0) && maxSize > 0 {
				params.Set("size", fmt.Sprint(maxSize))
			} else if s, err := strconv.Atoi(params["size"][0]); err != nil || maxSize < s {
				params.Set("size", fmt.Sprint(maxSize))
			}
			r.URL.RawQuery = params.Encode()
		}
		next.ServeHTTP(w, r)
	})
}
