package uifiles

import (
	"testing"
	"time"
)

func TestPrehashDataEncodeToken(t *testing.T) {

	key := []byte("testkey")

	// modt := time.Now()
	modt, _ := time.Parse(time.RFC3339, "2017-01-01T00:00:00Z")

	pd := PrehashData{}

	pd.Entries = append(pd.Entries, PrehashEntry{
		Name:    "js:example1",
		ModTime: modt,
	})
	pd.Entries = append(pd.Entries, PrehashEntry{
		Name:    "js:example2",
		ModTime: modt,
	})

	token, err := pd.EncodeToken(key)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("token: %s", token)

	pd2 := PrehashData{}
	err = pd2.DecodeToken(key, token)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("pd2 after decode: %#v", pd2)

	if len(pd2.Entries) != 2 {
		t.Fatalf("wrong length: %d", len(pd2.Entries))
	}

	if pd2.Entries[1].Name != "js:example2" || pd2.Entries[1].ModTime.UnixNano() != modt.UnixNano() {
		t.Logf("modt=%v", modt.UnixNano())
		t.Fatalf("bad pd2.Entries[1]: %q, %v", pd2.Entries[1].Name, pd2.Entries[1].ModTime.UnixNano())
	}

	if !pd.Equal(pd2) {
		t.Fatalf("pd and pd2 are not equal! (pd=%#v) (pd2=%#v)", pd, pd2)
	}

	if !pd.EqualIgnoreModTime(pd2) {
		t.Fatalf("pd and pd2 are not EqualIgnoreModTime! (pd=%#v) (pd2=%#v)", pd, pd2)
	}

	// change a time entry and check for equality again
	pd2.Entries[0].ModTime = time.Now()

	if pd.Equal(pd2) {
		t.Fatalf("pd and pd2 are equal after time change! (pd=%#v) (pd2=%#v)", pd, pd2)
	}

	if !pd.EqualIgnoreModTime(pd2) {
		t.Fatalf("pd and pd2 are not EqualIgnoreModTime! (pd=%#v) (pd2=%#v)", pd, pd2)
	}

}
