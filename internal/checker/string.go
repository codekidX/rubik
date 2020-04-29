package checker

import (
	"fmt"
	"strconv"
	"strings"
)

var stringAssertions = map[string]Assertion{
	"min-len":  strMin,
	"max-len":  strMax,
	"in-range": strRange,
}

func strMin(args ...string) (string, bool) {
	t := args[0]
	val := args[1]
	fmt.Println(t, val)
	opVal, err := strconv.Atoi(val)
	if err != nil || t == "" {
		return "Value passed as operator is not an integer", false
	}
	if len(t) < opVal {
		msg := fmt.Sprintf("%s must contain minimum %s characters", t, val)
		return msg, false
	}
	return "", true
}

func strMax(args ...string) (string, bool) {
	t := args[0]
	val := args[1]
	opVal, err := strconv.Atoi(val)
	if err != nil || t == "" {
		return "Value passed as operator is not an integer", false
	}
	if len(t) > opVal {
		msg := fmt.Sprintf("%s must contain maximum %s characters", t, val)
		return msg, false
	}
	return "", true
}

func strRange(args ...string) (string, bool) {
	t := args[0]
	val := args[1]
	if !strings.Contains(val, ">") || t == "" {
		return "VALIDATION: range assertion needs range operator: >", false
	}

	rge := strings.Split(val, ">")
	min, err := strconv.Atoi(rge[0])
	max, err := strconv.Atoi(rge[1])
	if err != nil {
		return "range operators should be a number/int not a string", false
	}

	if len(t) <= max && len(t) >= min {
		return "", true
	}
	return "", false
}
