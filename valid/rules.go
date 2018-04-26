package valid

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// QUESTIONS:
// Should a rule ever change a field?  Like can we make a "trim whitespace" or "convert to int" rule?
//  I'm leaning in the direction of yeah we should have that, but not certain it won't add too much
//  complexity
// Field names should probably all be in JSON format, even here in the Go code. And the various functions
//  that actually touch the fields can deal with the conversion but at least everything outside this
//  package will consistently use the external (JSON) representation whenever there's a string with
//  a field name in it.

// Rule is a validation rule which can be applied to an object.
// A Rule is responsible for knowing which field(s) it applies to and handling
// various types of objects including structs and maps - although that task
// should generally be delegated to ReadField() and WriteField().
// A Messages should be returned as the error if validation
// errors are found, or nil if validatin passes.  Other errors may be returned
// to address edge cases of invalid configuration but all validation problems
// must be expressed as a Messages.
type Rule interface {
	Apply(obj interface{}) error
}

type RuleFunc func(obj interface{}) error

func (f RuleFunc) Apply(obj interface{}) error {
	return f(obj)
}

type Rules []Rule

func (rl Rules) Apply(obj interface{}) error {
	var retmsgs Messages
	for _, r := range rl {
		err := r.Apply(obj)
		if err != nil {
			vml, ok := err.(Messages)
			if !ok {
				// other errors are just returned immediately
				return err
			}
			// validation messages are put together in one list
			retmsgs = append(retmsgs, vml...)
		}
	}
	if len(retmsgs) == 0 {
		return nil
	}
	// return message (or will be nil if non generated)
	return retmsgs
}

func DefaultFieldMessageName(fieldName string) string {

	var outParts []string

	parts := strings.Split(fieldName, "_")
	for _, p := range parts {
		if p == "" {
			continue
		}

		var p1, p2 string
		p1 = p[:1]
		p2 = p[1:]

		p = strings.ToUpper(p1) + p2

		outParts = append(outParts, p)

	}

	return strings.Join(outParts, " ")

}

// NewNotNilRule returns a rule that ensures a field is not nil.
func NewNotNilRule(fieldName string) Rule {
	return RuleFunc(func(obj interface{}) error {

		v, err := ReadField(obj, fieldName)
		if err != nil {
			return err
		}
		if v == nil {
			return Messages{Message{
				FieldName: fieldName,
				Code:      "notnil",
				Message:   fmt.Sprintf("%s is requried", DefaultFieldMessageName(fieldName)),
			}}
		}

		return nil

	})
}

// NewMinLenRule returns a rule that ensures the string representation of a
// field value is greater than or equal the specified number of bytes.
// This rule has no effect if the field value is nil.
func NewMinLenRule(fieldName string, minLen int) Rule {
	// log.Printf("NewMinLenRule(%q, %v)", fieldName, minLen)
	return RuleFunc(func(obj interface{}) error {
		v, err := ReadField(obj, fieldName)
		if err != nil {
			// log.Printf("NewMinLenRule, ReadField(%#v, %q) returned err %v", obj, fieldName, err)
			return err
		}
		if v == nil {
			return nil
		}

		var theErr error
		theErr = Messages{Message{
			FieldName: fieldName,
			Code:      "minlen",
			Message:   fmt.Sprintf("The minimum length for %s is %d", DefaultFieldMessageName(fieldName), minLen),
			Data:      map[string]interface{}{"value": minLen},
		}}

		vlen := len(fmt.Sprintf("%v", v))
		if vlen < minLen {
			return theErr
		}
		// validation succeeded
		return nil
	})
}

// NewMaxLenRule returns a rule that ensures the string representation of a
// field value is less than or equal the specified number of bytes.
// This rule has no effect if the field value is nil.
func NewMaxLenRule(fieldName string, minLen int) Rule {
	return RuleFunc(func(obj interface{}) error {
		v, err := ReadField(obj, fieldName)
		if err != nil {
			return err
		}
		if v == nil {
			return nil
		}

		var theErr error
		theErr = Messages{Message{
			FieldName: fieldName,
			Code:      "maxlen",
			Message:   fmt.Sprintf("The maximum length for %s is %d", DefaultFieldMessageName(fieldName), minLen),
			Data:      map[string]interface{}{"value": minLen},
		}}

		vlen := len(fmt.Sprintf("%v", v))
		if vlen > minLen {
			return theErr
		}
		// validation succeeded
		return nil
	})
}

// NewRegexpRule returns a rule that ensures the string representation of a
// field value matches the specified regexp pattern.
// This rule has no effect if the field value is nil.
func NewRegexpRule(fieldName string, pattern *regexp.Regexp) Rule {
	return RuleFunc(func(obj interface{}) error {
		v, err := ReadField(obj, fieldName)
		if err != nil {
			return err
		}
		if v == nil {
			return nil
		}

		vstr := fmt.Sprintf("%v", v)
		if !pattern.MatchString(vstr) {
			return Messages{Message{
				FieldName: fieldName,
				Code:      "regexp",
				Message:   fmt.Sprintf("%s does not match required pattern", DefaultFieldMessageName(fieldName)),
				Data:      map[string]interface{}{"regexp_string": pattern.String()},
			}}
		}

		// validation succeeded
		return nil
	})
}

const (
	EMAIL_REGEXP_PATTERN = `(?i)^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$`
)

// NewEmailRule returns a rule that ensures the string representation of a
// field value looks like an email address according to the pattern EMAIL_REGEXP_PATTERN.
// This rule has no effect if the field value is nil.
func NewEmailRule(fieldName string) Rule {

	// borrowed from: https://www.regular-expressions.info/email.html; I'm open to suggestions but looking
	// for something simple
	pattern := regexp.MustCompile(EMAIL_REGEXP_PATTERN)

	return RuleFunc(func(obj interface{}) error {
		v, err := ReadField(obj, fieldName)
		if err != nil {
			return err
		}
		if v == nil {
			return nil
		}

		vstr := fmt.Sprintf("%v", v)
		if !pattern.MatchString(vstr) {
			return Messages{Message{
				FieldName: fieldName,
				Code:      "email",
				Message:   fmt.Sprintf("%s does not appear to be a valid email address", DefaultFieldMessageName(fieldName)),
			}}
		}

		// validation succeeded
		return nil
	})
}

// NewMinValRule returns a rule that ensures the field value is equal to or greater than
// the value you specify.
// This rule has no effect if the field value is nil or if it is not numeric (integer or floating point).
func NewMinValRule(fieldName string, minval float64) Rule {

	// FIXME: should we also do something with NaN and infinity here?

	return RuleFunc(func(obj interface{}) error {
		v, err := ReadField(obj, fieldName)
		if err != nil {
			return err
		}
		if v == nil {
			return nil
		}

		var vfl float64

		vtype := reflect.TypeOf(v)
		// if it can't be converted, we just ignore the check
		if !vtype.ConvertibleTo(reflect.TypeOf(vfl)) {
			return nil
		}

		vval := reflect.ValueOf(v)
		reflect.ValueOf(&vfl).Elem().Set(vval.Convert(reflect.TypeOf(vfl)))

		if vfl < minval {
			return Messages{Message{
				FieldName: fieldName,
				Code:      "minval",
				Message:   fmt.Sprintf("%s is below the minimum value (%v)", DefaultFieldMessageName(fieldName), minval),
			}}
		}

		// validation succeeded
		return nil
	})
}

// NewMaxValRule returns a rule that ensures the field value is equal to or greater than
// the value you specify.
// This rule has no effect if the field value is nil or if it is not numeric (integer or floating point).
func NewMaxValRule(fieldName string, maxval float64) Rule {

	return RuleFunc(func(obj interface{}) error {
		v, err := ReadField(obj, fieldName)
		if err != nil {
			return err
		}
		if v == nil {
			return nil
		}

		var vfl float64

		vtype := reflect.TypeOf(v)
		// if it can't be converted, we just ignore the check
		if !vtype.ConvertibleTo(reflect.TypeOf(vfl)) {
			return nil
		}

		vval := reflect.ValueOf(v)
		reflect.ValueOf(&vfl).Elem().Set(vval.Convert(reflect.TypeOf(vfl)))

		if vfl > maxval {
			return Messages{Message{
				FieldName: fieldName,
				Code:      "maxval",
				Message:   fmt.Sprintf("%s is below the minimum value (%v)", DefaultFieldMessageName(fieldName), maxval),
			}}
		}

		// validation succeeded
		return nil
	})
}
