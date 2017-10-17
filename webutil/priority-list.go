package webutil

type PriorityListItem struct {
	Priority float64
	Name     string
	Value    interface{}
}

// PriorityList represents a list of items that each have a priority (by convention it's 0 to 100, but that's not enforced), a name and a value.
// All operations provided here are guaranteed to copy the original without modifying it.  Intended for use by global registries of
// things where you have a list of default things that builds up and it gets customized during application startup.
type PriorityList []PriorityListItem

// TODO: sort by priority, add/replace/find/remove by name, convert to []interface{}
