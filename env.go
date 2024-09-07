//go:build !tinygo

package thing2

import (
	"os"
)

func Getenv(name string, defaultValue string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	return value
}

func Setenv(name, value string) {
	os.Setenv(name, value)
}
