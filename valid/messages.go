package valid

import "fmt"

type Message struct {

	// indicates which object by name
	Object string `json:"object,omitempty"`

	// for multiple this is the index, starting with zero
	Index interface{} `json:"index,omitempty"`

	// name of the field
	FieldName string `json:"field_name,omitempty"`

	// the code indicating what validation rule failed
	Code interface{} `json:"code,omitempty"`

	// the English language message describing the failure
	Message string `json:"messsage,omitempty"`

	// the additional data (if any), which can be used to learn more about the validation failure or for i18n translation
	Data interface{} `json:"data,omitempty"`
}

type Messages []Message

func (ms Messages) Error() string {
	return fmt.Sprintf("%d validation message(s)", len(ms))
	// b, err := json.Marshal(ms)
	// if err != nil {
	// 	return err.Error()
	// }
	// return string(b)
}

func (ms Messages) ContainsCode(code string) bool {
	for _, m := range ms {
		if m.Code == code {
			return true
		}
	}
	return false
}
