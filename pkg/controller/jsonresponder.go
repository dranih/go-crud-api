package controller

import (
	"log"
	"net/http"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/dranih/go-crud-api/pkg/record"
)

type JsonResponder struct {
	debug bool
	rf    *ResponseFactory
}

func NewJsonResponder(debug bool) *JsonResponder {
	return &JsonResponder{debug, &ResponseFactory{}}
}

func (jr *JsonResponder) Error(errorCode int, argument string, w http.ResponseWriter, r *http.Request, details interface{}) http.ResponseWriter {
	document := record.NewErrorDocument(record.NewErrorCode(errorCode), argument, details)
	return jr.rf.FromObject(document.GetStatus(), document, w, r)
}

func (jr *JsonResponder) Success(result interface{}, w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	return jr.rf.FromObject(record.OK, result, w, r)
}

func (jr *JsonResponder) Exception(err error, w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	document := record.NewErrorDocumentFromError(err, jr.debug)
	if jr.debug {
		addExceptionHeaders(w, err)
		log.Printf("Error : %s", err.Error())
	}
	response := jr.rf.FromObject(document.GetStatus(), document, w, r)
	return response
}

func (jr *JsonResponder) Multi(results *[]interface{}, errs []error, w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	success := true
	documents := []interface{}{}
	errors := []record.ErrorDocument{}
	for i, result := range *results {
		if errs[i] != nil {
			documents = append(documents, nil)
			errors = append(errors, *record.NewErrorDocumentFromError(errs[i], jr.debug))
			success = false
			if jr.debug {
				addExceptionHeaders(w, errs[i])
			}
		} else {
			documents = append(documents, result)
			errors = append(errors, *record.NewErrorDocument(record.NewErrorCode(0), "", ""))
		}
	}
	var response http.ResponseWriter
	if !success {
		response = jr.rf.FromObject(record.FAILED_DEPENDENCY, errors, w, r)
	} else {
		response = jr.rf.FromObject(record.OK, documents, w, r)
	}
	return response
}

func (jr *JsonResponder) SetAfterHandler(handler string) error {
	if jr.rf.afterHandler == nil {
		if handler != "" {
			if t, err := template.New("handler").Funcs(sprig.TxtFuncMap()).Parse(handler); err == nil {
				jr.rf.afterHandler = t
				return nil
			} else {
				log.Printf("Error : could not parse template beforeHandler : %s", err.Error())
				return err
			}
		}
	}
	return nil
}
