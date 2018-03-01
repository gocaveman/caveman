package autowire

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type T1 struct {
	NoTag string // no tag
	T2V   *T2    `autowire:""`                     // empty tag
	T3V   *T3    `autowire:"t3"`                   // named tag
	T4V   T4     `autowire:"t4"`                   // interface
	Opt   string `autowire:"nothinghere,optional"` // optional
}

var afterWireCalled bool = false

func (t1 *T1) AfterWire() error {
	log.Printf("AfterWire here")
	afterWireCalled = true
	return nil
}

type T2 struct{}

type T3 struct{}

type T4 interface {
	Error() string
}

func TestAutoWire(t *testing.T) {

	assert := assert.New(t)

	t1 := &T1{}
	t2 := &T2{}
	t3 := &T3{}
	t4 := fmt.Errorf("test")

	w := &Wirer{}
	w.Populate(t1)
	w.Provide("", t2)
	w.Provide("t3", t3)
	w.Provide("t4", t4)

	err := w.Run()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("TestAutoWire result: %#v", t1)

	assert.Equal("", t1.NoTag)
	assert.Equal(t2, t1.T2V)
	assert.Equal(t3, t1.T3V)
	assert.Equal(t4, t1.T4V)
	assert.Equal("", t1.Opt)

	assert.True(afterWireCalled)

	// TODO: check for a valid that is required but missing

}
