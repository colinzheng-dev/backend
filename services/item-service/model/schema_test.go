package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadSchemas(t *testing.T) {
	schemas, _ := readRawSchemas()
	expandSchemas(schemas)
	hasExtends := false
	for _, s := range schemas {
		_, chk := s["extends"]
		hasExtends = hasExtends || chk
	}
	assert.False(t, hasExtends)
}
