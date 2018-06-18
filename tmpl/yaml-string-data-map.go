package tmpl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/gocaveman/caveman/webutil"
	yaml "gopkg.in/yaml.v2"
)

// // StringDataMap describes a map of string keys and generic interface values.
// // This interface matches StringDataMap in the pageinfo package, so the meta data
// // is interoperable.  Implementations are not thread-safe.
// type StringDataMap interface {
// 	Value(key string) interface{}
// 	Keys() []string
// 	Set(key string, val interface{}) // Set("key", nil) will delete "key"
// }

// type check
var _ webutil.StringDataMap = &YAMLStringDataMap{}

// YAMLStringDataMap implements StringDataMap supporting a subset of YAML and
// facilitates reading and writing while attempting to preserve comments and sequence.
// It is intended for use in the top section of a template file (see FileSystemStore).
type YAMLStringDataMap struct {
	entries         []YAMLMapEntry
	entryMap        map[string]int
	trailingComment string
}

// Map returns a new map[string]interface{} with the data from this instance.
func (m *YAMLStringDataMap) Map() map[string]interface{} {
	if m == nil {
		return nil
	}
	ret := make(map[string]interface{}, len(m.entries))
	for _, e := range m.entries {
		ret[e.Key] = e.Value
	}
	return ret
}

func (m *YAMLStringDataMap) rebuildEntryMap() {
	m.entryMap = make(map[string]int, len(m.entries))
	for i, e := range m.entries {
		m.entryMap[e.Key] = i
	}

}

func (m *YAMLStringDataMap) Data(key string) interface{} {
	i, ok := m.entryMap[key]
	if !ok {
		return nil
	}
	return m.entries[i].Value
}

func (m *YAMLStringDataMap) Keys() (ret []string) {
	for _, e := range m.entries {
		ret = append(ret, e.Key)
	}
	return
}

func (m *YAMLStringDataMap) Set(key string, val interface{}) {

	i, ok := m.entryMap[key]
	if !ok {

		if val == nil { // deleting non-existent key, nop
			return
		}

		// create a new entry and append it
		var e YAMLMapEntry
		e.Key = key
		e.Value = val
		m.entryMap[key] = len(m.entries)
		m.entries = append(m.entries, e)

		return
	}

	// update value for existing entry

	// delete case
	if val == nil {

		es := m.entries
		es[i].Value = nil
		copy(es[i:], es[i+1:])
		m.entries = es[:len(es)-1]

		m.rebuildEntryMap()
		return
	}

	// set case is a simple assignment
	m.entries[i].Value = val
	return
}

type YAMLMapEntry struct {
	Comment string
	Key     string
	Value   interface{}
}

// NewYAMLStringDataMap returns a newly initialized YAMLStringDataMap.
func NewYAMLStringDataMap() *YAMLStringDataMap {
	log.Printf("FIXME: YAML is probably a mistake, consider switching to TOML (whitespace sensitivity is a bad thing)")
	return &YAMLStringDataMap{
		entryMap: make(map[string]int),
	}
}

// BUG(bgp): Keep an eye on this: https://github.com/go-yaml/yaml/pull/219 - would be better
// to use a (more) correctly implemented comment parsing solution.

// ReadYAMLStringDataMap reads input that is a subset of YAML.
// The data must be a YAML map and comments above each individual map entry
// are preserved, as is the sequence of the keys.
func ReadYAMLStringDataMap(in io.Reader) (*YAMLStringDataMap, error) {

	ret := NewYAMLStringDataMap()

	br := bufio.NewReader(in)

	var commentBuf bytes.Buffer // the current comment we are building
	var yamlBuf bytes.Buffer    // the current yaml map entry we are building
	var inComment = true        // are we currently reading the comment

	flushEntry := func() error {
		defer func() {
			commentBuf.Truncate(0)
			yamlBuf.Truncate(0)
			inComment = true
		}()

		ms := yaml.MapSlice{}
		err := yaml.Unmarshal(yamlBuf.Bytes(), &ms)
		if err != nil {
			return err
		}

		for _, mi := range ms {

			// ugly, but basically we needed to use MapSlice above to preserve the order,
			// but now we have to marshal and unmarshal again using map[string]interface{},
			// so we don't have these yaml library internal types in our values
			b, err := yaml.Marshal(yaml.MapSlice{mi})
			if err != nil {
				return err
			}
			var mi2 map[string]interface{}
			err = yaml.Unmarshal(b, &mi2)
			if err != nil {
				return err
			}

			var e YAMLMapEntry
			keyStr, ok := mi.Key.(string)
			if !ok {
				return fmt.Errorf("YAMLStringDataMap does not support key of type %T (val=%v)", mi.Key, mi.Key)
			}
			e.Key = keyStr
			e.Value = mi2[keyStr] // use the value from mi2
			if commentBuf.Len() > 0 {
				e.Comment = commentBuf.String()
				commentBuf.Truncate(0)
			}
			ret.entries = append(ret.entries, e)
		}

		return nil
	}

	// read line by line
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		lineTrimmed := strings.TrimSpace(line)
		lineIsComment := lineTrimmed == "" || lineTrimmed[0] == '#'

		switch {

		// reading comments, continue
		case inComment && lineIsComment:
			commentBuf.WriteString(line)

		// reading yaml entry, continue
		case !inComment && !lineIsComment:
			yamlBuf.WriteString(line)

		// were in comment but now have yaml entry, switch to yaml entry
		case inComment && !lineIsComment:
			inComment = false
			yamlBuf.WriteString(line)

		// were reading yaml entry but now have comment
		case !inComment && lineIsComment:
			err := flushEntry() // save the entry, clear bufs and set inComment = true
			if err != nil {
				return nil, err
			}
			commentBuf.WriteString(line)

		}

	}

	if yamlBuf.Len() > 0 {
		err := flushEntry()
		if err != nil {
			return nil, err
		}
	} else if commentBuf.Len() > 0 {
		ret.trailingComment = commentBuf.String()
	}

	ret.rebuildEntryMap()

	return ret, nil
}

// WriteYAMLStringDataMap will write a YAMLStringDataMap, preserving sequence and comments
// where preserved from ReadYAMLStringDataMap.
func WriteYAMLStringDataMap(out io.Writer, m *YAMLStringDataMap) error {

	for _, e := range m.entries {

		if e.Comment != "" {
			_, err := fmt.Fprint(out, e.Comment)
			if err != nil {
				return err
			}
			if !strings.HasSuffix(e.Comment, "\n") {
				_, err := out.Write([]byte("\n"))
				if err != nil {
					return err
				}
			}
		}

		ms := yaml.MapSlice{
			yaml.MapItem{
				Key:   e.Key,
				Value: e.Value,
			},
		}

		b, err := yaml.Marshal(&ms)
		if err != nil {
			return err
		}

		_, err = out.Write(b)
		if err != nil {
			return err
		}

		if !bytes.HasSuffix(b, []byte("\n")) {

			_, err = out.Write([]byte("\n"))
			if err != nil {
				return err
			}

		}

	}

	if len(m.trailingComment) > 0 {
		_, err := out.Write([]byte(m.trailingComment))
		if err != nil {
			return err
		}
	}

	return nil
}
