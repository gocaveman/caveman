package dbutil

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringObjMap is map[string]interface{} that encodes to JSON in the database.
type StringObjMap map[string]interface{}

// MarshalJSON customizes the marshal so that a nil map encodes to "{}" not "null".
// This makes things more consistent for JS clients and easier to work with.
func (m StringObjMap) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte(`{}`), nil
	}
	return json.Marshal(map[string]interface{}(m))
}

func (m StringObjMap) Value() (driver.Value, error) {
	if m == nil {
		m = StringObjMap{}
	}
	b, err := json.Marshal(m)
	return driver.Value(b), err
}

func (m *StringObjMap) Scan(value interface{}) error {

	var ret StringObjMap
	var b []byte

	if value != nil {

		switch v := value.(type) {
		case []byte:
			b = v
		case string:
			b = []byte(v)
		default:
			return fmt.Errorf("cannot convert from StringObjMap to sql driver type %T", value)
		}

		err := json.Unmarshal(b, &ret)
		if err != nil {
			return err
		}

	} else {
		ret = StringObjMap{}
	}

	// ensure we always have a valid map
	if ret == nil {
		ret = StringObjMap{}
	}

	*m = ret

	return nil
}
