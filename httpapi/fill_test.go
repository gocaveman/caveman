package httpapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FillS1 struct {
	F1 string      `httpapi:"f1,fillable"`
	F2 string      `httpapi:"f2"`
	F3 interface{} `httpapi:"f3,fillable"`
	F4 int         `httpapi:"f4,fillable"`
	F5 string      `httpapi:",fillable"`
}

func TestFill(t *testing.T) {

	assert := assert.New(t)

	// simple map<-map case
	{
		m1 := map[string]string{
			"k1": "v1",
			"k2": "v2",
		}

		m2 := make(map[string]interface{})

		assert.NoError(Fill(m2, m1))
		assert.Equal("v1", m1["k1"])
		assert.Equal("v2", m2["k2"])
	}

	// now try map<-struct
	{
		m1 := map[string]interface{}{
			"f1": "v1",
			"f2": "v2",
			"f3": "v3",
			"f4": 4,
			"f5": "v5",
		}

		s1 := FillS1{}

		assert.NoError(Fill(&s1, m1))
		assert.Equal("v1", s1.F1)
		assert.Equal("", s1.F2) // not fillable
		assert.Equal("v3", s1.F3)
		assert.Equal(4, s1.F4)
		assert.Equal("v5", s1.F5)
	}

	// now struct<-struct
	{
		ssrc := &FillS1{
			F1: "v1",
			F2: "v2",
			F3: "v3",
			F4: 4,
			F5: "v5",
		}

		s1 := FillS1{}

		assert.NoError(Fill(&s1, ssrc))
		assert.Equal("v1", s1.F1)
		assert.Equal("", s1.F2) // not fillable
		assert.Equal("v3", s1.F3)
		assert.Equal(4, s1.F4)
		assert.Equal("v5", s1.F5)
	}

}
