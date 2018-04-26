package valid

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// SUGGESTED STRUCT TAG EXAMPLES SHOWING FORMAT:
//
// Field1 string      `valid:"required,minlen=6"`
// Field2 interface{} `valid:"type=int,minval=10,maxval=1000"`

// type StructTagRuleFunc func(fieldName string, part string) (Rules, error)

// // StructTagInfo is a package global var so if an application wants to register
// // it's own struct tag parsing rules it can
// var StructTagInfo []StructTagRuleFunc

var DefaultTagRuleGenerators []TagRuleGenerator

func init() {

	DefaultTagRuleGenerators = append(DefaultTagRuleGenerators, TagRuleGeneratorFunc(func(t reflect.Type, goFieldName string, tagValues url.Values) (Rules, error) {
		if len(tagValues["notnil"]) > 0 {
			fieldName := MakeRuleFieldName(goFieldName, tagValues)
			return Rules{NewNotNilRule(fieldName)}, nil
		}
		return nil, nil
	}))

	DefaultTagRuleGenerators = append(DefaultTagRuleGenerators, TagRuleGeneratorFunc(func(t reflect.Type, goFieldName string, tagValues url.Values) (Rules, error) {
		i, _ := strconv.Atoi(tagValues.Get("minlen"))
		if i > 0 {
			fieldName := MakeRuleFieldName(goFieldName, tagValues)
			return Rules{NewMinLenRule(fieldName, i)}, nil
		}
		return nil, nil
	}))

	DefaultTagRuleGenerators = append(DefaultTagRuleGenerators, TagRuleGeneratorFunc(func(t reflect.Type, goFieldName string, tagValues url.Values) (Rules, error) {
		i, _ := strconv.Atoi(tagValues.Get("maxlen"))
		if i > 0 {
			fieldName := MakeRuleFieldName(goFieldName, tagValues)
			return Rules{NewMaxLenRule(fieldName, i)}, nil
		}
		return nil, nil
	}))

	DefaultTagRuleGenerators = append(DefaultTagRuleGenerators, TagRuleGeneratorFunc(func(t reflect.Type, goFieldName string, tagValues url.Values) (Rules, error) {
		re := tagValues.Get("regexp")
		if re != "" {
			fieldName := MakeRuleFieldName(goFieldName, tagValues)
			rec, err := regexp.Compile(re)
			if err != nil {
				return nil, err
			}
			return Rules{NewRegexpRule(fieldName, rec)}, nil
		}
		return nil, nil
	}))

	DefaultTagRuleGenerators = append(DefaultTagRuleGenerators, TagRuleGeneratorFunc(func(t reflect.Type, goFieldName string, tagValues url.Values) (Rules, error) {
		_, ok := tagValues["email"]
		if ok {
			fieldName := MakeRuleFieldName(goFieldName, tagValues)
			return Rules{NewEmailRule(fieldName)}, nil
		}
		return nil, nil
	}))

	DefaultTagRuleGenerators = append(DefaultTagRuleGenerators, TagRuleGeneratorFunc(func(t reflect.Type, goFieldName string, tagValues url.Values) (Rules, error) {

		mvstr, ok := tagValues["minval"]
		if ok && len(mvstr) > 0 {

			var mv float64
			_, err := fmt.Sscanf(mvstr[0], "%f", &mv)
			if err != nil {
				return nil, fmt.Errorf("error parsing minval %q: %v", mvstr[0], err)
			}
			fieldName := MakeRuleFieldName(goFieldName, tagValues)
			return Rules{NewMinValRule(fieldName, mv)}, nil

		}

		return nil, nil
	}))

	DefaultTagRuleGenerators = append(DefaultTagRuleGenerators, TagRuleGeneratorFunc(func(t reflect.Type, goFieldName string, tagValues url.Values) (Rules, error) {

		mvstr, ok := tagValues["maxval"]
		if ok && len(mvstr) > 0 {

			var mv float64
			_, err := fmt.Sscanf(mvstr[0], "%f", &mv)
			if err != nil {
				return nil, fmt.Errorf("error parsing maxval %q: %v", mvstr[0], err)
			}
			fieldName := MakeRuleFieldName(goFieldName, tagValues)
			return Rules{NewMaxValRule(fieldName, mv)}, nil

		}

		return nil, nil
	}))

}

type TagRuleGenerator interface {
	// Generate one or more rules from the Go struct field name, the name inside
	// the 'valid' part of the struct tag and it's value(s).
	TagRuleGenerate(t reflect.Type, goFieldName string, tagValues url.Values) (Rules, error)
}

type TagRuleGeneratorFunc func(t reflect.Type, goFieldName string, tagValues url.Values) (Rules, error)

func (f TagRuleGeneratorFunc) TagRuleGenerate(t reflect.Type, goFieldName string, tagValues url.Values) (Rules, error) {
	return f(t, goFieldName, tagValues)
}

func StructRules(t reflect.Type, gens []TagRuleGenerator) (Rules, error) {

	if gens == nil {
		gens = DefaultTagRuleGenerators
	}

	var rules Rules

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		vstr := sf.Tag.Get("valid")
		vvalues := StructTagToValues(vstr)

		for _, gen := range gens {
			rs, err := gen.TagRuleGenerate(t, sf.Name, vvalues)
			if err != nil {
				return nil, err
			}
			rules = append(rules, rs...)
		}

	}

	return rules, nil
}

func MakeRuleFieldName(goFieldName string, tagValues url.Values) string {
	name := tagValues.Get("name")
	if name != "" {
		return name
	}
	ret := goFieldName // use the go field name as-is
	// ret := ToSnake(goFieldName)
	// log.Printf("MakeRuleFieldName(goFieldName=%q) is returning %q", goFieldName, ret)
	return ret
}

func StructTagToValues(st string) url.Values {

	ret := make(url.Values)

	parts := strings.Split(st, ",")

	for _, part := range parts {
		kvparts := strings.SplitN(part, "=", 2)
		if len(kvparts) < 2 {
			ret.Set(kvparts[0], "")
		} else {
			ret.Set(kvparts[0], kvparts[1])
		}
	}

	return ret
}
