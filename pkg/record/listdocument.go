package record

import (
	"encoding/json"
	"log"
)

type ListDocument struct {
	records []map[string]interface{}
	results int
}

func NewListDocument(records []map[string]interface{}, results int) *ListDocument {
	return &ListDocument{records, results}
}

func (l *ListDocument) GetRecords() []map[string]interface{} {
	return l.records
}

func (l *ListDocument) GetResults() int {
	return l.results
}

func (l *ListDocument) Serialize() map[string]interface{} {
	return map[string]interface{}{"records": l.records, "results": l.results}
}

func (l *ListDocument) JsonSerialize() string {
	/*return array_filter($this->serialize(), function ($v) {
		return $v !== -1;
	});*/
	data, err := json.Marshal(l.Serialize())
	if err != nil {
		log.Printf("Marshaling error : %v", err)
	}
	return string(data)
}
