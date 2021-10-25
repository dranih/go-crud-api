package controller

import "net/http"

type JsonResponder struct {
	debug bool
	rf    *ResponseFactory
}

func NewJsonResponder(debug bool) *JsonResponder {
	return &JsonResponder{debug, &ResponseFactory{}}
}

// not finished (errordocument)
func (jr *JsonResponder) Error(errorCode int, argument interface{}, w http.ResponseWriter, details ...interface{}) http.ResponseWriter {
	return jr.rf.FromObject(NOT_FOUND, argument, w)
}

/*

public function error(int $error, string $argument, $details = null): ResponseInterface
{
	$document = new ErrorDocument(new ErrorCode($error), $argument, $details);
	return ResponseFactory::fromObject($document->getStatus(), $document);
}
*/
func (jr *JsonResponder) Success(result interface{}, w http.ResponseWriter) http.ResponseWriter {
	return jr.rf.FromObject(OK, result, w)
}

// not finished (errordocument)
func (jr *JsonResponder) Exception(err error, w http.ResponseWriter) http.ResponseWriter {
	return jr.rf.FromObject(NOT_FOUND, "", w)
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
	return jr.rf.FromObject(OK, results, w)
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
