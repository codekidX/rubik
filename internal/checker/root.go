package checker

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Assertion is the assert functions for rubiks validation cycle
type Assertion func(...string) (string, bool)

// Check validation string
func Check(target reflect.Value, vstring string) error {
	cleaned := strings.TrimSpace(vstring)
	ops := strings.Split(cleaned, ",")

	switch target.Kind() {
	case reflect.String:
		for _, op := range ops {
			return assert(stringAssertions, op)
		}
		break
	case reflect.Int:
		break
	}
	return nil
}

func assert(assMap map[string]Assertion, op string) error {
	if assMap != nil {
		if strings.Contains(op, "=") {
			args := strings.Split(op, "=")
			msg, ok := assMap[op](args...)
			if !ok {
				return errors.New(msg)
			}
		}
	}
	return nil
}
