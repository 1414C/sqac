package common

import (
	"regexp"
	"strings"
)

// CamelToSnake converts camelCase to snake_case
func CamelToSnake(s string) string {

	// create a capture group for the string (.)
	// create a second capture group ([A-Z][a-z]+)
	// [A-Z] match any character in the set
	// [a-z] match any characrer in the set
	// + match one or more of the preceding token
	// group find the first UpperCase letter [A-Z] followed by any number
	// of LowerCase letters [a-z]+.
	// oneCamel - match: 'eCamel'
	// testCamelCaseIBMPowerEdge - match: {'lCamel', 'MPower'}
	var matchCapLc = regexp.MustCompile("(.)([A-Z][a-z]+)")

	// ([a-z0-9]) capture group for all characters in the prescribed ranges
	// ([A-Z]) capture group for all characters in the prescribed range
	// testCamelCaseIBMPowerEdge - match: {'tC', 'lC', 'eI', 'Re'}
	var matchLcCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	// replace all found lc->uc and lc->uc->uc with lc_uc
	// in the source string.
	// 'eCamel' -> 'e_Camel'
	// testCamelCaseIBMPowerEdge -> 'test_CamelCaseIBM_PowerEdge'
	sc := matchCapLc.ReplaceAllString(s, "${1}_${2}")

	// test_CamelCaseIBM_PowerEdge -> 'test_Camel_Case_IBM_Power_Edge'
	sc = matchLcCap.ReplaceAllString(sc, "${1}_${2}")
	return strings.ToLower(sc)
}
