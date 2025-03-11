package auth

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestUUID(t *testing.T) {
	for i := 0; i < 36; i++ {
		uuid := uuid.New().String()
		fmt.Println(len([]rune(uuid)))
	}
}