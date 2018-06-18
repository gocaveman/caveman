package renderer

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
)

// DEFAULT_FUNCMAP is the FuncMap used if not overridden. (See New() function.)
// Generally, it is not recommended to change the FuncMap.  Prefer assigning
// objects to the context as this is much less likely to cause dependency problems
// and namespace issues.  The default FuncMap is really only here to provide easy
// access to functionality that aides in template rendering in the most general way
// and is common to virtualy all applications.
var DEFAULT_FUNCMAP template.FuncMap = template.FuncMap{
	// "CallJSON": func(ctx interface{}) interface{} { return ctx }, // FIXME: need to figure this out...
	// TODO: probably HTML, CSS, JS, look around and see if anything else super common

	// TODO: we definitely need something that sets one or more context keys and
	// returns a new context

	// First will return the first non-nil value you pass it.
	"First": func(o ...interface{}) interface{} {
		for _, ov := range o {
			if ov != nil {
				return ov
			}
		}
		return nil
	},

	// WithValue calls and returns the result of context.WithValue(ctx, key, val).
	"WithValue": func(ctx context.Context, key string, val interface{}) context.Context {
		return context.WithValue(ctx, key, val)
	},

	// WithValueMapJSON takes a string unmarshals as JSON to map[string]interface{} and assigns
	// each entry to the context.
	"WithValueMapJSON": func(ctx context.Context, jsonStr string) (context.Context, error) {
		m := make(map[interface{}]interface{}, 4)
		err := json.Unmarshal([]byte(jsonStr), &m)
		if err != nil {
			return ctx, err
		}
		return WithValueMap(ctx, m), nil
	},
}

func WithValueMap(ctx context.Context, values map[interface{}]interface{}) context.Context {
	return &valueMapCtx{Context: ctx, valueMap: values}
}

type valueMapCtx struct {
	context.Context
	valueMap map[interface{}]interface{}
}

func (c *valueMapCtx) String() string {
	return fmt.Sprintf("%v.WithValueMap(%#v)", c.Context, c.valueMap)
}

func (c *valueMapCtx) Value(key interface{}) interface{} {
	v, ok := c.valueMap[key]
	if ok {
		return v
	}
	return c.Context.Value(key)
}

// TODO: it probably makes sense to take some common functionality from the Go stdlib and
// expose it to templates with sane naming. For example:
// NewStringsHandler().... results in a "strings" context value which has functions on it
// like HasPrefix(), Split(), Join(), ToLower(), etc.
// FIGURE OUT WHERE THIS SHOULD GO!  I'M NOT SURE THAT renderer IS THE RIGHT PLACE FOR IT,
// ALTHOUGH THERE'S A STRONG ARGUMENT FOR THAT. (MAYBE "renderutil")
