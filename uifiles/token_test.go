package uifiles

import (
	"testing"
)

func TestPrehashDataEncodeToken(t *testing.T) {

	// assert := assert.New(t)

	key := []byte("testkey")

	pd := []string{
		"js:example1",
		"js:example2",
	}

	token, err := EncodeToken(pd, key)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("token: %s", token)

	var pd2 []string
	err = DecodeToken(&pd2, key, token)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("pd2 after decode: %#v", pd2)

	if len(pd2) != 2 {
		t.Fatalf("wrong length: %d", len(pd2))
	}

	if pd2[1] != "js:example2" {
		t.Fatalf("bad pd2[1]: %q", pd2[1])
	}

}
