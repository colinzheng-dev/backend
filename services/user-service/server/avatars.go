package server

import (
	"fmt"
	"math/rand"
)

func generateAvatar(format string, count int) func() string {
	return func() string {
		return fmt.Sprintf(format, rand.Intn(count)+1)
	}
}
