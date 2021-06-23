package mailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexp(t *testing.T) {
	tests := []struct {
		full  string
		email string
		name  string
	}{
		{`Veganbase Login <login@veganbase.com>`,
			`login@veganbase.com`,
			`Veganbase Login`,
		},
	}
	for _, test := range tests {
		res := fromRE.FindStringSubmatch(test.full)
		assert.Len(t, res, 3)
		assert.Equal(t, test.email, res[2])
		assert.Equal(t, test.name, res[1])
	}
}
