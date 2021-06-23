package server

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParams(t *testing.T) {
	tests := []struct {
		u  string
		ok bool
	}{
		{"type=hotel", true},
		{"type=bad", false},
		{"type=hotel&page=3&per_page=10", true},
		{"type=hotel&page=x&per_page=10", false},
		{"type=hotel&page=1&per_page=abc", false},
		{"type=restaurant&q=london", true},
		{"type=rest&q=london", false},
		{"user=usr_DFS3rdfSF4sdf&approval=pending", true},
		{"user=usr_DFS3rdfSF4sdf&approval=pend", false},
		{"geo=51.2,10.3&dist=100", true},
		{"geo=51.2,10.3", false},
		{"geo=51.2&dist=100", false},
	}
	for _, test := range tests {
		url, err := url.Parse("http://staging.veganapi.com/items?" + test.u)
		assert.Nil(t, err)
		r := http.Request{URL: url}
		_, err = Params(&r, true)
		if test.ok {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}
}
