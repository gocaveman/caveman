package dbutil

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringValueList is just a string slice but has SQL marshaling to/from JSON.
// JSON values are always an empty array ("[]") and never "null".  Inputs
// which are nil will be converted to an emtpy array automatically.
type StringValueList []string

func (l StringValueList) Value() (driver.Value, error) {
	if l == nil {
		l = StringValueList{}
	}
	b, err := json.Marshal(l)
	return driver.Value(b), err
}

func (l *StringValueList) Scan(value interface{}) error {

	var ret StringValueList
	var b []byte

	if value != nil {

		switch v := value.(type) {
		case []byte:
			b = v
		case string:
			b = []byte(v)
		default:
			return fmt.Errorf("cannot convert from StringValueList to sql driver type %T", value)
		}

		err := json.Unmarshal(b, &ret)
		if err != nil {
			return err
		}

	} else {
		ret = StringValueList{}
	}

	*l = ret

	return nil
}

func (l StringValueList) String() string {
	dv, err := l.Value()
	if err != nil {
		return fmt.Sprintf("(string conversion resulted in error: %v)", err)
	}
	return fmt.Sprintf("%s", dv)
}

func (l StringValueList) Contains(v string) bool {
	for _, sv := range l {
		if sv == v {
			return true
		}
	}
	return false
}

func (l StringValueList) IndexOf(v string) int {
	for i, sv := range l {
		if sv == v {
			return i
		}
	}
	return -1
}

// RemoveFast deletes the value at the specified position
// without preserving sequence and returns the new slice.
// Note that the array backing the original slice is modified.
func (l StringValueList) RemoveFast(i int) StringValueList {
	// swap value we're removing to end of array
	l[len(l)-1], l[i] = l[i], l[len(l)-1]
	// return with last element omitted
	return l[:len(l)-1]
}
