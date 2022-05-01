package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
)

type ReconnectMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
	db         *database.GenericDB
}

func NewReconnectMiddleware(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService, db *database.GenericDB) *ReconnectMiddleware {
	return &ReconnectMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection, db: db}
}

func (rm *ReconnectMiddleware) getStringFromHandler(handlerStr string) string {
	handler := fmt.Sprint(rm.getProperty(handlerStr, ""))
	if handler != "" {
		if t, err := template.New(handlerStr).Funcs(sprig.TxtFuncMap()).Parse(handler); err == nil {
			var res bytes.Buffer
			if err := t.Execute(&res, nil); err == nil {
				if res.String() != "" {
					return res.String()
				}
			} else {
				log.Printf("Error : could not execute template %s : %s", handlerStr, err.Error())
			}
		} else {
			log.Printf("Error : could not parse template %s : %s", handlerStr, err.Error())
		}
	}
	return ""
}

func (rm *ReconnectMiddleware) getIntFromHandler(handlerStr string) int {
	handler := fmt.Sprint(rm.getProperty(handlerStr, ""))
	if handler != "" {
		if t, err := template.New(handlerStr).Funcs(sprig.TxtFuncMap()).Parse(handler); err == nil {
			var res bytes.Buffer
			if err := t.Execute(&res, nil); err == nil {
				if i, err := strconv.Atoi(res.String()); err != nil {
					return i
				}
			} else {
				log.Printf("Error : could not execute template %s : %s", handlerStr, err.Error())
			}
		} else {
			log.Printf("Error : could not parse template %s : %s", handlerStr, err.Error())
		}
	}
	return 0
}

func (rm *ReconnectMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		driver := rm.getStringFromHandler("driverHandler")
		address := rm.getStringFromHandler("addressHandler")
		port := rm.getIntFromHandler("portHandler")
		database := rm.getStringFromHandler("databaseHandler")
		//We expect tables in a csv string and convert to map[string]bool
		var tables map[string]bool
		if tablesStr := rm.getStringFromHandler("tablesHandler"); tablesStr != "" {
			for _, table := range strings.Split(tablesStr, ",") {
				tables[table] = true
			}
		}
		mapping := rm.getMapProperty("mappingHandler", "")
		username := rm.getStringFromHandler("usernameHandler")
		password := rm.getStringFromHandler("passwordHandler")
		if driver != "" || address != "" || port > 0 || database != "" || tables != nil || mapping != nil || username != "" || password != "" {
			rm.db.Reconstruct(driver, address, port, database, tables, mapping, username, password)
		}
		next.ServeHTTP(w, r)
	})
}
