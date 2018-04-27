package httpapi

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type S1 struct {
	F1 string      `json:"f1"`
	F2 []string    `json:"f2"`
	F3 interface{} `json:"f3"`
	F4 string      `json:"f4"`
	F5 int         `json:"f5"`
	F6 string      `json:"-"`
	F7 []byte      `json:"f7"`
	F8 float64     `json:"f8"`
	F9 string      `json:"f9"`
}

func TestFormUnmarshal(t *testing.T) {

	assert := assert.New(t)

	// try map[string]string
	m1 := make(map[string]string)
	assert.NoError(FormUnmarshal(mustParseQuery("k1=v1&k2=va&k2=vb&k3*=vA"), &m1))
	assert.Equal("v1", m1["k1"])
	assert.Equal("va", m1["k2"])
	assert.Equal("vA", m1["k3"])

	// try map[string][]string
	m2 := make(map[string][]string)
	assert.NoError(FormUnmarshal(mustParseQuery("k1=v1&k2=va&k2=vb&k3*=vA"), &m2))
	assert.Equal("v1", m2["k1"][0])
	assert.Equal("va", m2["k2"][0])
	assert.Equal("vb", m2["k2"][1])
	assert.Equal("vA", m2["k3"][0])

	// try map[string]interface{}
	m3 := make(map[string]interface{})
	assert.NoError(FormUnmarshal(mustParseQuery("k1=v1&k2=va&k2=vb&k3*=vA"), &m3))
	assert.Equal("v1", m3["k1"])
	assert.Equal("va", m3["k2"])
	assert.Equal([]interface{}{"vA"}, m3["k3"])

	// try some funky characters (null, control, quotes, newline Unicode BMP, Unicode non-BMP [emojis])
	m4 := make(map[string]string)
	funkyStr := "FUNK:\000\n\t\"'☂☕👍😂"
	assert.NoError(FormUnmarshal(mustParseQuery("test1="+funkyStr), &m4))
	assert.Equal(funkyStr, m4["test1"])

	s1a := &S1{}
	assert.NoError(FormUnmarshal(mustParseQuery("f1=blah&f2=a&f2=b&f3=A&f3=B&f4=&f5=10&f6=test&f7=dGVzdA%3D%3D&f8=3.141592654&f9*=A"), &s1a))
	assert.Equal("blah", s1a.F1)
	assert.Equal("a", s1a.F2[0])
	assert.Equal("b", s1a.F2[1])
	assert.Equal("A", s1a.F3)
	assert.Equal("", s1a.F4)
	assert.Equal(10, s1a.F5)
	assert.Equal("", s1a.F6)
	assert.Equal([]byte("test"), s1a.F7)
	assert.Equal(float64(3.141592654), s1a.F8)
	assert.Equal("A", s1a.F9)

	s1a = &S1{}
	b, err := FormToJSON(mustParseQuery("f1=blah&f2=a&f2=b&f3=A&f3=B&f4=&f5=10&f6=test&f7=dGVzdA%3D%3D&f8=3.141592654"), &s1a)
	assert.NoError(err)
	t.Logf("b=%s", b)

}

func mustParseQuery(q string) url.Values {

	vals, err := url.ParseQuery(q)
	if err != nil {
		panic(err)
	}

	return vals
}
