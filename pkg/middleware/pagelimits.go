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
			maxPageInt := 100
			maxPage := pm.getProperty("pages", maxPageInt)
			switch v := maxPage.(type) {
			case string:
				if a, err := strconv.Atoi(v); err != nil {
					maxPageInt = a
				}
			case int:
				maxPageInt = v
			}

			if v, ok := params["page"]; ok && len(v) > 0 && maxPageInt > 0 {
				var page int
				var err error
				if strings.Index(v[0], ",") == -1 {
					page, err = strconv.Atoi(v[0])
				} else {
					page, err = strconv.Atoi(strings.SplitN(v[0], ",", 2)[0])
				}
				if err == nil && page > maxPageInt {
					pm.Responder.Error(record.PAGINATION_FORBIDDEN, "", w, nil)
					return
				}
			}

			maxSizeInt := 1000
			maxSize := pm.getProperty("records", maxSizeInt)
			switch v := maxSize.(type) {
			case string:
				if a, err := strconv.Atoi(v); err != nil {
					maxSizeInt = a
				}
			case int:
				maxSizeInt = v
			}

			if v, ok := params["size"]; (!ok || len(v) == 0) && maxSizeInt > 0 {
				params.Set("size", fmt.Sprint(maxSizeInt))
			} else if s, err := strconv.Atoi(params["size"][0]); err != nil || maxSizeInt < s {
				params.Set("size", fmt.Sprint(maxSizeInt))
			}
			r.URL.RawQuery = params.Encode()
		}
		next.ServeHTTP(w, r)
	})
}
