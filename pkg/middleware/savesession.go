package middleware

import (
	"net/http"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type SaveSessionMiddleware struct {
	GenericMiddleware
}

func NewSaveSession(responder controller.Responder, properties map[string]interface{}) *SaveSessionMiddleware {
	return &SaveSessionMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}}
}

func (ssm *SaveSessionMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := utils.GetSession(w, r)
		if err := session.Save(r, w); err != nil {
			ssm.Responder.Error(record.INTERNAL_SERVER_ERROR, err.Error(), w, "")
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
