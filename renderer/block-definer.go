package renderer

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"github.com/gocaveman/caveman/webutil/handlerregistry"
)

const RendererBlockDefiner = "renderer.BlockDefiner"

func init() {
	handlerregistry.MustRegister(handlerregistry.SeqSetup, "BlockDefineHandler", BlockDefineHandler{})
}

// BlockDefineHandler is a handler which makes a BlockDefiner available in the request context
// with the key "renderer.BlockDefiner".
type BlockDefineHandler struct{}

func NewBlockDefineHandler() BlockDefineHandler {
	return BlockDefineHandler{}
}

func (h BlockDefineHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	ctx := r.Context()
	// don't override existing one
	if ctx.Value(RendererBlockDefiner) != nil {
		return w, r
	}
	ctx = context.WithValue(ctx, RendererBlockDefiner, &blockDefiner{})
	r = r.WithContext(ctx)
	return w, r
}

// BlockDefiner is the interface packages can use to manually define blocks to be used during the render process.
// The name is the template name, content is the html/template contnent string and condition is
// a function that can be used to veto inclusion based on a context.  If condition is nil it is
// the same as providing a function that always returns true.
type BlockDefiner interface {
	BlockDefine(name, content string, condition func(ctx context.Context) (bool, error))
}

type BlockMapper interface {
	BlockMap(ctx context.Context) (map[string]string, error)
}

type blockDefinerValue struct {
	content   string
	condition func(ctx context.Context) (bool, error)
}

type blockDefiner struct {
	blocks map[string]blockDefinerValue
}

func (b *blockDefiner) BlockDefine(name, content string, condition func(ctx context.Context) (bool, error)) {
	if b.blocks == nil {
		b.blocks = make(map[string]blockDefinerValue)
	}
	b.blocks[name] = blockDefinerValue{content: content, condition: condition}
}
func (b *blockDefiner) BlockMap(ctx context.Context) (map[string]string, error) {
	ret := make(map[string]string, len(b.blocks))
	for k, v := range b.blocks {
		if v.condition != nil {
			result, err := v.condition(ctx)
			if err != nil {
				return ret, fmt.Errorf("BlockMap checking condition for %q resulted in error: %v", k, err)
			}
			if !result {
				continue
			}
		}
		ret[k] = v.content
	}
	return ret, nil
}

// NewBlockDefineModifier returns a modifier which looks for a BlockMapper and uses it to define blocks.
// This works in conjunction with BlockDefineHandler which does the setup for Go code to be able to manually define blocks.
func NewBlockDefineModifier() TemplateModifier {
	return TemplateModifierFunc(func(ctx context.Context, t *template.Template) (context.Context, *template.Template, error) {

		// ensure we're set up and we have a non-empty block map
		blockMapper, ok := ctx.Value(RendererBlockDefiner).(BlockMapper)
		if !ok {
			// log.Printf("BlockDefineModifier bailed, no block mapper")
			return ctx, t, nil
		}
		blockMap, err := blockMapper.BlockMap(ctx)
		if err != nil {
			return ctx, t, err
		}
		if len(blockMap) == 0 {
			// log.Printf("BlockDefineModifier bailed, block mapper empty")
			return ctx, t, nil
		}

		// sort the keys and do it in sequence
		keys := make([]string, 0, len(blockMap))
		for k := range blockMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		// log.Printf("BlockDefineModifier running with %d entries", len(keys))
		// create and parse each template
		for _, k := range keys {
			t2 := t.New(k)
			t3, err := t2.Parse(blockMap[k])
			if err != nil {
				return ctx, t, err
			}
			t = t3
		}

		return ctx, t, nil
	})
}
