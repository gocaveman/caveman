package webutil

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringDataMap describes a map of string keys and generic interface values.
// Implementations are not thread-safe.
type StringDataMap interface {
	Data(key string) interface{}
	Keys() []string
	Set(key string, val interface{}) // Set("key", nil) will delete "key"
}

// StringDataMapReader has read methods for a map, is a subset of StringDataMap.
type StringDataMapReader interface {
	Data(key string) interface{}
	Keys() []string
}

// type check
var _ StringDataMap = SimpleStringDataMap{}

// SimpleStringDataMap implements StringDataMap on top of map[string]interface{}.
type SimpleStringDataMap map[string]interface{}

func (m SimpleStringDataMap) Data(key string) interface{} {
	return m[key]
}
func (m SimpleStringDataMap) Keys() (ret []string) {
	for k := range m {
		ret = append(ret, k)
	}
	return
}
func (m SimpleStringDataMap) Set(key string, val interface{}) {
	if val == nil {
		delete(m, key)
	} else {
		m[key] = val
	}
}

func (m *SimpleStringDataMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	*m = make(SimpleStringDataMap)
	err := fmt.Errorf("cannot Scan() type %T", value)
	switch v := value.(type) {
	case string:
		err = json.Unmarshal([]byte(v), m)
	case []byte:
		err = json.Unmarshal(v, m)
	}
	return err
}

func (m SimpleStringDataMap) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	return driver.Value(string(b)), err
}
