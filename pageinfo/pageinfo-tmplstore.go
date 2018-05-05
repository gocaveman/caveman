package pageinfo

// TemplateStore is a subset of tmpl.Store of the things we need.
type TemplateStore interface {
	ReadTemplate(category, fileName string) (body []byte, mimeType string, meta map[string]interface{}, err error)
	FindByPrefix(category, fileNamePrefix string, limit int) ([]string, error)
}

// PageInfoFromTmplStore adapts a tmpl.Store to a pageinfo.Store with certain rules applied.
type PageInfoFromTmplStore struct {
	TemplateStore TemplateStore
}
