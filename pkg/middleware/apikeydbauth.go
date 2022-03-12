package middleware

import (
	"fmt"
	"net/http"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type ApiKeyDbAuthMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
	db         *database.GenericDB
	ordering   *database.OrderingInfo
}

func NewApiKeyDbAuth(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService, db *database.GenericDB) *ApiKeyDbAuthMiddleware {
	return &ApiKeyDbAuthMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection, db: db, ordering: database.NewOrderingInfo()}
}

func (akdam *ApiKeyDbAuthMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := map[string]interface{}{}
		ok := true
		headerName := fmt.Sprint(akdam.getProperty("header", "X-API-Key"))
		apiKey := r.Header.Get(headerName)
		if apiKey != "" {
			tableName := fmt.Sprint(akdam.getProperty("usersTable", "users"))
			table := akdam.reflection.GetTable(tableName)
			apiKeyColumnName := fmt.Sprint(akdam.getProperty("apiKeyColumn", "api_key"))
			apiKeyColumn := table.GetColumn(apiKeyColumnName)
			condition := database.NewColumnCondition(apiKeyColumn, "eq", apiKey)
			columnNames := table.GetColumnNames()
			columnOrdering := akdam.ordering.GetDefaultColumnOrdering(table)
			users := akdam.db.SelectAll(table, columnNames, condition, columnOrdering, 0, 1)
			if len(users) < 1 {
				akdam.Responder.Error(record.AUTHENTICATION_FAILED, apiKey, w, "")
				ok = false
			} else {
				user = users[0]
			}
		} else {
			if authenticationMode := akdam.getProperty("mode", "required"); authenticationMode == "required" {
				realm := fmt.Sprint(akdam.getProperty("realm", "Api key required"))
				w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
				akdam.Responder.Error(record.AUTHENTICATION_REQUIRED, "", w, realm)
				ok = false
			}
		}
		if ok {
			if apiKey != "" {
				session := utils.GetSession(w, r)
				session.Values["apiUser"] = user
			}
			next.ServeHTTP(w, r)
		}
	})
}
