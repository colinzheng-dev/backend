package chassis

import (
	"strings"
	"testing"
)

func TestIDGeneration(t *testing.T) {
	ids := make(map[string]bool, 0)
	for i := 0; i < 1000; i++ {
		ids[NewID("ofr")] = true
	}

	if len(ids) != 1000 {
		t.Error("non-unique IDs")
	}

	if !strings.HasPrefix(NewID("htl"), "htl_") {
		t.Error("ID prefix generation wrong")
	}
}
