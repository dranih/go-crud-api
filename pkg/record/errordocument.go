package record

import (
	"encoding/json"
)

type ErrorDocument struct {
	errorCode *ErrorCode
	argument  string
	details   string
}

func NewErrorDocument(errorCode *ErrorCode, argument, details string) *ErrorDocument {
	return &ErrorDocument{errorCode, argument, details}
}

func (ed *ErrorDocument) GetStatus() int {
	return ed.errorCode.GetStatus()
}

func (ed *ErrorDocument) GetCode() int {
	return ed.errorCode.GetCode()
}

func (ed *ErrorDocument) GetMessage() string {
	return ed.errorCode.GetMessage(ed.argument)
}

func (ed *ErrorDocument) Serialize() map[string]interface{} {
	return map[string]interface{}{"code": ed.GetCode(),
		"message": ed.GetMessage(),
		"details": ed.details}
}

// json marshaling for struct ErrorDocument
func (ed *ErrorDocument) MarshalJSON() ([]byte, error) {
	return json.Marshal(ed.Serialize())
}

/*
public static function fromException(\Throwable $exception, bool $debug)
{
	$document = new ErrorDocument(new ErrorCode(ErrorCode::ERROR_NOT_FOUND), $exception->getMessage(), null);
	if ($exception instanceof \PDOException) {
		if (strpos(strtolower($exception->getMessage()), 'duplicate') !== false) {
			$document = new ErrorDocument(new ErrorCode(ErrorCode::DUPLICATE_KEY_EXCEPTION), '', null);
		} elseif (strpos(strtolower($exception->getMessage()), 'unique constraint') !== false) {
			$document = new ErrorDocument(new ErrorCode(ErrorCode::DUPLICATE_KEY_EXCEPTION), '', null);
		} elseif (strpos(strtolower($exception->getMessage()), 'default value') !== false) {
			$document = new ErrorDocument(new ErrorCode(ErrorCode::DATA_INTEGRITY_VIOLATION), '', null);
		} elseif (strpos(strtolower($exception->getMessage()), 'allow nulls') !== false) {
			$document = new ErrorDocument(new ErrorCode(ErrorCode::DATA_INTEGRITY_VIOLATION), '', null);
		} elseif (strpos(strtolower($exception->getMessage()), 'constraint') !== false) {
			$document = new ErrorDocument(new ErrorCode(ErrorCode::DATA_INTEGRITY_VIOLATION), '', null);
		} else {
			$message = $debug?$exception->getMessage():'PDOException occurred (enable debug mode)';
			$document = new ErrorDocument(new ErrorCode(ErrorCode::ERROR_NOT_FOUND), $message, null);
		}
	}
	return $document;
}
*/
