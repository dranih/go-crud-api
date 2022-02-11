package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type CustomizationMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
}

func NewCustomizationMiddleware(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService) *CustomizationMiddleware {
	return &CustomizationMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection}
}

func (cm *CustomizationMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operation := utils.GetOperation(r)
		tableName := utils.GetPathSegment(r, 2)
		env := map[string]interface{}{}

		beforeHandler := fmt.Sprint(cm.getProperty("beforeHandler", ""))
		if beforeHandler != "" {
			if t, err := template.New("beforeHandler").Funcs(sprig.TxtFuncMap()).Parse(beforeHandler); err == nil {
				data := struct {
					Operation   string
					TableName   string
					Request     *http.Request
					Environment map[string]interface{}
				}{Operation: operation, TableName: tableName, Request: r, Environment: env}
				var res bytes.Buffer
				if err := t.Execute(&res, data); err == nil {
					env["tablename"] = tableName
					env["operation"] = operation
					session := utils.GetSession(w, r)
					session.Values["environment"] = env
				} else {
					log.Printf("Error : could not execute template beforeHandler : %s", err.Error())
				}
			} else {
				log.Printf("Error : could not parse template beforeHandler : %s", err.Error())
			}
		}

		//Passing the after handler to the response factory
		afterHandler := fmt.Sprint(cm.getProperty("afterHandler", ""))
		if afterHandler != "" {
			if err := cm.Responder.SetAfterHandler(afterHandler); err != nil {
				log.Printf("Error : could not parse template beforeHandler : %s", err.Error())
			}
		}

		next.ServeHTTP(w, r)
	})
}
