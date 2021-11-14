package controller

import (
	"net/http"

	"github.com/dranih/go-crud-api/pkg/record"
)

type JsonResponder struct {
	debug bool
	rf    *ResponseFactory
}

func NewJsonResponder(debug bool) *JsonResponder {
	return &JsonResponder{debug, &ResponseFactory{}}
}

func (jr *JsonResponder) Error(errorCode int, argument string, w http.ResponseWriter, details string) http.ResponseWriter {
	document := record.NewErrorDocument(record.NewErrorCode(errorCode), argument, details)
	return jr.rf.FromObject(document.GetStatus(), document, w)
}

func (jr *JsonResponder) Success(result interface{}, w http.ResponseWriter) http.ResponseWriter {
	return jr.rf.FromObject(record.OK, result, w)
}

func (jr *JsonResponder) Exception(err error, w http.ResponseWriter) http.ResponseWriter {
	document := record.NewErrorDocumentFromError(err, jr.debug)
	if jr.debug {
		addExceptionHeaders(w, err)
	}
	response := jr.rf.FromObject(document.GetStatus(), document, w)
	return response
}

func (jr *JsonResponder) Multi(results *[]map[string]interface{}, errs []error, w http.ResponseWriter) http.ResponseWriter {
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
		response = jr.rf.FromObject(record.FAILED_DEPENDENCY, errors, w)
	} else {
		response = jr.rf.FromObject(record.OK, documents, w)
	}
	return response
}
