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

// not finished (errordocument)
func (jr *JsonResponder) Exception(err error, w http.ResponseWriter) http.ResponseWriter {
	return jr.rf.FromObject(record.NOT_FOUND, "", w)
}

/*
public function exception($exception): ResponseInterface
{
	$document = ErrorDocument::fromException($exception, $this->debug);
	$response = ResponseFactory::fromObject($document->getStatus(), $document);
	if ($this->debug) {
		$response = ResponseUtils::addExceptionHeaders($response, $exception);
	}
	return $response;
}

*/
// not finished (errordocument)
func (jr *JsonResponder) Multi(results interface{}, w http.ResponseWriter) http.ResponseWriter {
	return jr.rf.FromObject(record.OK, results, w)
}

/*
public function multi($results): ResponseInterface
{
	$documents = array();
	$errors = array();
	$success = true;
	foreach ($results as $i => $result) {
		if ($result instanceof \Throwable) {
			$documents[$i] = null;
			$errors[$i] = ErrorDocument::fromException($result, $this->debug);
			$success = false;
		} else {
			$documents[$i] = $result;
			$errors[$i] = new ErrorDocument(new ErrorCode(0), '', null);
		}
	}
	$status = $success ? ResponseFactory::OK : ResponseFactory::FAILED_DEPENDENCY;
	$document = $success ? $documents : $errors;
	$response = ResponseFactory::fromObject($status, $document);
	foreach ($results as $i => $result) {
		if ($result instanceof \Throwable) {
			if ($this->debug) {
				$response = ResponseUtils::addExceptionHeaders($response, $result);
			}
		}
	}
	return $response;
}
*/
