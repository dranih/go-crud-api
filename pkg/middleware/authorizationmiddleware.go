package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/utils"

	"text/template"
)

type AuthorizationMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
}

func NewAuthorizationMiddleware(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection}
}

func (am *AuthorizationMiddleware) handleColumns(operation, tableName string) {
	columnHandler := fmt.Sprint(am.getProperty("columnHandler", ""))
	if columnHandler != "" {
		table := am.reflection.GetTable(tableName)
		if t, err := template.New("columnHandler").Funcs(sprig.TxtFuncMap()).Parse(columnHandler); err == nil {
			for _, columnName := range table.GetColumnNames() {
				var res bytes.Buffer
				data := struct {
					Operation  string
					ColumnName string
				}{Operation: operation, ColumnName: columnName}
				if err := t.Execute(&res, data); err == nil {
					if allowed, _ := strconv.ParseBool(strings.TrimSpace(res.String())); !allowed {
						table.RemoveColumn(columnName)
					}
				} else {
					log.Printf("Error : could not execute template tableHandler : %s", err.Error())
				}
			}
		} else {
			log.Printf("Error : could not parse template columnHandler : %s", err.Error())
		}
	}
}

func (am *AuthorizationMiddleware) handleTable(operation, tableName string) {
	if !am.reflection.HasTable(tableName) {
		return
	}
	allowed := true
	tableHandler := fmt.Sprint(am.getProperty("tableHandler", ""))
	if tableHandler != "" {
		if t, err := template.New("tableHandler").Funcs(sprig.TxtFuncMap()).Parse(tableHandler); err == nil {
			var res bytes.Buffer
			data := struct {
				Operation string
				TableName string
			}{Operation: operation, TableName: tableName}
			if err := t.Execute(&res, data); err == nil {
				allowed, _ = strconv.ParseBool(strings.TrimSpace(res.String()))
			} else {
				log.Printf("Error : could not execute template tableHandler : %s", err.Error())
			}
		} else {
			log.Printf("Error : could not parse template tableHandler : %s", err.Error())
		}
	}
	if !allowed {
		am.reflection.RemoveTable(tableName)
	} else {
		am.handleColumns(operation, tableName)
	}
}

func (am *AuthorizationMiddleware) handleRecords(operation, tableName string) {
	if !am.reflection.HasTable(tableName) {
		return
	}
	recordHandler := fmt.Sprint(am.getProperty("recordHandler", ""))
	if recordHandler != "" {
		if t, err := template.New("recordHandler").Funcs(sprig.TxtFuncMap()).Parse(recordHandler); err == nil {
			var res bytes.Buffer
			data := struct {
				Operation string
				TableName string
			}{Operation: operation, TableName: tableName}
			if err := t.Execute(&res, data); err == nil {
				query := strings.TrimSpace(res.String())
				filters := &database.FilterInfo{}
				table := am.reflection.GetTable(tableName)
				query = strings.Replace(strings.Replace(query, "=", "[]=", -1), "][]=", "]=", -1)
				if params, err := url.ParseQuery(query); err == nil {
					condition := filters.GetCombinedConditions(table, params)
					utils.VStore.Set(fmt.Sprintf("authorization.conditions.%s", tableName), condition)
				} else {
					log.Printf("Error : parse recordHandler query : %s", err.Error())
				}
			} else {
				log.Printf("Error : could not execute template recordHandler : %s", err.Error())
			}
		} else {
			log.Printf("Error : could not parse template recordHandler : %s", err.Error())
		}
	}
}

func (am *AuthorizationMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := utils.GetPathSegment(r, 1)
		operation := utils.GetOperation(r)
		tableNames := utils.GetTableNames(r, am.reflection.GetTableNames())
		for _, tableName := range tableNames {
			am.handleTable(operation, tableName)
			if path == "records" {
				am.handleRecords(operation, tableName)
			}
		}
		if path == "openapi" {
			utils.VStore.Set("authorization.tableHandler", am.getProperty("tableHandler", ""))
			utils.VStore.Set("authorization.columnHandler", am.getProperty("columnHandler", ""))
		}
		next.ServeHTTP(w, r)
	})
}
