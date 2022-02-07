package middleware

import (
	"net/http"
	"strings"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type JoinLimitsMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
}

func NewJoinLimitsMiddleware(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService) *JoinLimitsMiddleware {
	return &JoinLimitsMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection}
}

func (jm *JoinLimitsMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operation := utils.GetOperation(r)
		params := utils.GetRequestParams(r)
		joins, ok := params["join"]
		if ok && len(joins) > 0 && (operation == "list" || operation == "read") {
			maxDepth := jm.getIntProperty("depth", 3)
			maxTables := jm.getIntProperty("tables", 10)
			maxRecords := jm.getIntProperty("records", 1000)
			tableCount := 0
			joinPaths := []string{}
			for i := 0; i < len(joins); i++ {
				joinPath := []string{}
				tables := strings.Split(joins[i], ",")
				for depth := 0; depth < utils.MinInt(maxDepth, len(tables)); depth++ {
					joinPath = append(joinPath, tables[depth])
					tableCount += 1
					if tableCount == maxTables {
						break
					}
				}
				joinPaths = append(joinPaths, strings.Join(joinPath, ","))
				if tableCount == maxTables {
					break
				}
			}
			params["join"] = joinPaths
			r.URL.RawQuery = params.Encode()
			utils.VStore.Set("joinLimits.maxRecords", maxRecords)
		}
		next.ServeHTTP(w, r)
	})
}
