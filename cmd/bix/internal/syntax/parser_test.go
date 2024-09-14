// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package syntax

import (
	"errors"
	"fmt"
	"testing"
)

func TestEmptyParse(t *testing.T) {
	res, errs := Parse("", "parser_test.go")
	if errs != nil {
		t.Fatalf("%d errors occurred. Showing first error: %v", len(errs), errs[0])
	}
	if res != nil {
		t.Fatalf("got %v, want %v", res, nil)
	}
}

func parseErrorTestRunner(t *testing.T, s string, want string) {
	got, errs := Parse(s, "parser_test.go")
	if errs == nil {
		t.Errorf("No error occurred, got %q, want error %q", got, want)
	} else if len(errs) != 1 {
		t.Errorf("got %d errors, want 1 error", len(errs))

		for i, err := range errs {
			t.Errorf("error #%d: got error %q", i+1, errors.Unwrap(err))
		}
	} else if errors.Unwrap(errs[0]).Error() != want {
		t.Errorf("got error %q, want error %q", errors.Unwrap(errs[0]), want)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"\\", "unexpected eof after '\\'"},
		{"\\ ", "stray character after '\\'"},
		{"1 +", "expression expected"},
		{"1 + \\\n", "expression expected"},
		{"let", "expected variable name"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			parseErrorTestRunner(t, tt.src, tt.want)
		})
	}
}
