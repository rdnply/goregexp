package goregexp

import (
	"fmt"
	"testing"
)

func TestMatchingFirstExpr(t *testing.T) {
	type testCase struct {
		Name     string
		In       string
		Expected bool
	}

	testCases := []testCase{
		{Name: "Common test 1", In: "bcd", Expected: true},
		{Name: "Common test 2", In: "aabcdcd", Expected: true},
		{Name: "Common test 3", In: "aaaaaaaaaaaaaaaaacdcdcdcd", Expected: true},
		{Name: "Common test 4", In: "", Expected: false},
		{Name: "Common test 5", In: "aaabcdcd", Expected: true},
		{Name: "Common test 6", In: "aaabbbcdcd", Expected: false},
		{Name: "Common test 7", In: "aaaaaaaaaab", Expected: false},
		{Name: "Common test 8", In: "aaaaaaa", Expected: false},
		{Name: "Common test 9", In: "rrrrrrrrrrrr", Expected: false},
	}

	match, err := CreateMatcher("a*b?(cd)+")
	if err != nil {
		t.Fatal("incorrect regular expression")
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := match(tc.In)
			if actual != tc.Expected {
				t.Errorf("Match(%+v) = %+v, want: %+v", tc.In, actual, tc.Expected)
			}
		})
	}
}

func TestMatchingSecondExpr(t *testing.T) {
	type testCase struct {
		Name     string
		In       string
		Expected bool
	}

	testCases := []testCase{
		{Name: "Common test 1", In: "aaaaaaaaaaa", Expected: true},
		{Name: "Common test 2", In: "aabababababb", Expected: true},
		{Name: "Common test 3", In: "d", Expected: true},
		{Name: "Common test 4", In: "bbbb", Expected: true},
		{Name: "Common test 5", In: "", Expected: false},
		{Name: "Common test 6", In: "aaaaaaaaaaaaaad", Expected: false},
	}

	match, err := CreateMatcher("(((a|b)+)|d)")
	if err != nil {
		t.Fatal("incorrect regular expression")
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := match(tc.In)
			if actual != tc.Expected {
				t.Errorf("Match(%+v) = %+v, want: %+v", tc.In, actual, tc.Expected)
			}
		})
	}
}

func TestCreateMatcher(t *testing.T) {
	type testCase struct {
		Name     string
		In       string
		Expected error
	}

	openBracketErr := fmt.Errorf("missing open bracket in expression")
	closeBracketErr := fmt.Errorf("missing close bracket in expression")

	testCases := []testCase{
		{Name: "Common test 1", In: "(((a|b)+)|d)", Expected: nil},
		{Name: "Common test 2", In: "a*b+(d)?", Expected: nil},
		{Name: "Missing closing bracket", In: "(ab*", Expected: closeBracketErr},
		{Name: "Missing closing bracket", In: "(ab)*(a|b", Expected: closeBracketErr},
		{Name: "Missing open bracket 1", In: "ab)", Expected: openBracketErr},
		{Name: "Missing open bracket 2", In: "(ab)*A|B)", Expected: openBracketErr},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, actual := CreateMatcher(tc.In)
			if actual != nil {
				var expected string
				if tc.Expected == nil {
					expected = ""
				} else {
					expected = tc.Expected.Error()
				}

				if actual.Error() != expected {
					t.Errorf("CreateMatcher(%+v) = %+v, want: %+v", tc.In, actual, tc.Expected)
				}
			} else {
				if actual != tc.Expected {
					t.Errorf("CreateMatcher(%+v) = %+v, want: %+v", tc.In, actual, tc.Expected)
				}
			}

		})
	}
}
