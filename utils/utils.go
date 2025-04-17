package utils

import (
	"crypto/rand"
	"fmt"
)

func GenerateTraceCode() string {
	b := make([]byte, 2)
	rand.Read(b)
return fmt.Sprintf("%04X", b)
}
