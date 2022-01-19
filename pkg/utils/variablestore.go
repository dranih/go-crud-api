package utils

type VariableStore struct {
	values map[string]interface{}
}

var VStore = &VariableStore{}

func (vs *VariableStore) Get(key string) interface{} {
	res, exists := vs.values[key]
	if exists {
		return res
	}
	return nil
}

func (vs *VariableStore) Set(key string, value interface{}) {
	vs.values[key] = value
}
