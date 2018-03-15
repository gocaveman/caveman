package webutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicAuthHandler(t *testing.T) {

	assert := assert.New(t)

	h := NewBasicAuthOneHandler("joe", "secret")
	hl := NewDefaultHandlerList(h, http.NotFoundHandler())

	s := httptest.NewServer(hl)
	defer s.Close()

	client := s.Client()

	res, err := client.Get(s.URL)
	assert.NoError(err)
	assert.Equal(401, res.StatusCode)

	req, _ := http.NewRequest("GET", s.URL, nil)
	req.SetBasicAuth("joe", "wrong password")
	res, err = client.Do(req)
	assert.NoError(err)
	assert.Equal(401, res.StatusCode)

	req, _ = http.NewRequest("GET", s.URL, nil)
	req.SetBasicAuth("joe", "secret")
	res, err = client.Do(req)
	assert.NoError(err)
	assert.Equal(404, res.StatusCode) // auth should work and it should fall through to 404

}
