package i18n

import (
	"bytes"
	"html/template"
	"strings"
)

// Process provides a simple and fast way to execute Go template rendering logic in an i18n string.
// TODO: examples of use-cases
type Processor interface {
	Process(s string, ctx interface{}) (string, error)
}

// NewDefaultProcessor makes a new instance with default settings.
func NewDefaultProcessor() *DefaultProcessor {
	return &DefaultProcessor{}
}

// DefaultProcessor is the default implementation of the Processor interface.
type DefaultProcessor struct{}

func (p *DefaultProcessor) Process(s string, ctx interface{}) (string, error) {

	// if no template code in string just return as-is
	if !strings.Contains(s, "{{") {
		return s, nil
	}

	// TODO: LRU cache?

	var err error
	t := template.New("_")
	t, err = t.Parse(s)
	if err != nil {
		return s, err
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "_", ctx)
	if err != nil {
		return s, err
	}

	return buf.String(), nil
}
