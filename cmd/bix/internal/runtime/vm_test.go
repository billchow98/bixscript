// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package runtime

import (
	"errors"
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"log"
	"strings"
	"testing"
)

type TestVm struct {
	*Vm
	stringBuilder *strings.Builder
}

func newTestVm() TestVm {
	// TODO: This is for real logging to file in the future if needed
	//log.SetFlags(log.LstdFlags | log.Lshortfile)
	t := TestVm{Vm: New(), stringBuilder: new(strings.Builder)}
	t.SetLogger(log.New(t.stringBuilder, "", 0))
	return t
}

func (t *TestVm) Run(s string, file string) (string, []error) {
	errs := t.Vm.Run(s, file)
	s = t.stringBuilder.String()
	if len(s) != 0 {
		assert.Assert(s[len(s)-1] == '\n')
		s = s[:len(s)-1]
	}
	return s, errs
}

func RegressionTestRunner(t *testing.T, s string, want string) {
	v := newTestVm()
	got, errs := v.Run(s, "vm_test.go")
	if errs != nil {
		t.Fatalf("%d errors occurred. Showing first error: %v", len(errs), errs[0])
	}
	if got != want {
		t.Errorf("got %q, want %q\n", got, want)
	}
}

func TestRegressions(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"", ""},
		{"1", ""},
		{"print(1)\n", "1"},
		{"function fb() {\n}", ""},
		{"function call(f) {\n    f()\n}\nfunction foo() {\n    print(\"foo\")\n}\ncall(foo)", "foo"},
		{"print(1/2)\n", "0.5"},
		{"function foo() { return }", ""},
		{"function foo() {\n    return\n}", ""},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			RegressionTestRunner(t, tt.src, tt.want)
		})
	}
}

func ErrorTestRunner(t *testing.T, s string, want string) {
	v := newTestVm()
	got, errs := v.Run(s, "vm_test.go")
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

func TestErrors(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"0 + false", "addition arguments must both be Numbers or both be Strings"},
		{"false + 0", "addition arguments must both be Numbers or both be Strings"},
		{"false + false", "addition arguments must both be Numbers or both be Strings"},
		{`0 + ""`, "addition arguments must both be Numbers or both be Strings"},
		{`"" + 0`, "addition arguments must both be Numbers or both be Strings"},
		{`false + ""`, "addition arguments must both be Numbers or both be Strings"},
		{`"" + false`, "addition arguments must both be Numbers or both be Strings"},
		{"0 - false", "both subtraction arguments must be Numbers"},
		{"false- 0", "both subtraction arguments must be Numbers"},
		{"false - false", "both subtraction arguments must be Numbers"},
		{`0 - ""`, "both subtraction arguments must be Numbers"},
		{`"" - 0`, "both subtraction arguments must be Numbers"},
		{`false - ""`, "both subtraction arguments must be Numbers"},
		{`"" - false`, "both subtraction arguments must be Numbers"},
		{"0 * false", "both multiplication arguments must be Numbers"},
		{"false * 0", "both multiplication arguments must be Numbers"},
		{"false * false", "both multiplication arguments must be Numbers"},
		{`0 * ""`, "both multiplication arguments must be Numbers"},
		{`"" * 0`, "both multiplication arguments must be Numbers"},
		{`false * ""`, "both multiplication arguments must be Numbers"},
		{`"" * false`, "both multiplication arguments must be Numbers"},
		{"0 / false", "both division arguments must be Numbers"},
		{"false / 0", "both division arguments must be Numbers"},
		{"false / false", "both division arguments must be Numbers"},
		{`0 / ""`, "both division arguments must be Numbers"},
		{`"" / 0`, "both division arguments must be Numbers"},
		{`false / ""`, "both division arguments must be Numbers"},
		{`"" / false`, "both division arguments must be Numbers"},
		{"-false", "negation argument must be a Number"},
		{`-""`, "negation argument must be a Number"},
		{"!0", "logical not argument must be a Boolean"},
		{`!""`, "logical not argument must be a Boolean"},
		{"0 < false", "both comparison arguments must be Numbers"},
		{"false < 0", "both comparison arguments must be Numbers"},
		{"false < false", "both comparison arguments must be Numbers"},
		{`0 < ""`, "both comparison arguments must be Numbers"},
		{`"" < 0`, "both comparison arguments must be Numbers"},
		{`false < ""`, "both comparison arguments must be Numbers"},
		{`"" < false`, "both comparison arguments must be Numbers"},
		{"0 <= false", "both comparison arguments must be Numbers"},
		{"false <= 0", "both comparison arguments must be Numbers"},
		{"false <= false", "both comparison arguments must be Numbers"},
		{`0 <= ""`, "both comparison arguments must be Numbers"},
		{`"" <= 0`, "both comparison arguments must be Numbers"},
		{`false <= ""`, "both comparison arguments must be Numbers"},
		{`"" <= false`, "both comparison arguments must be Numbers"},
		{"0 > false", "both comparison arguments must be Numbers"},
		{"false > 0", "both comparison arguments must be Numbers"},
		{"false > false", "both comparison arguments must be Numbers"},
		{`0 > ""`, "both comparison arguments must be Numbers"},
		{`"" > 0`, "both comparison arguments must be Numbers"},
		{`false > ""`, "both comparison arguments must be Numbers"},
		{`"" > false`, "both comparison arguments must be Numbers"},
		{"0 >= false", "both comparison arguments must be Numbers"},
		{"false >= 0", "both comparison arguments must be Numbers"},
		{"false >= false", "both comparison arguments must be Numbers"},
		{`0 >= ""`, "both comparison arguments must be Numbers"},
		{`"" >= 0`, "both comparison arguments must be Numbers"},
		{`false >= ""`, "both comparison arguments must be Numbers"},
		{`"" >= false`, "both comparison arguments must be Numbers"},
		{"print(foo)", "use of undeclared variable"},
		{"let foo = 123\nfoo = false", "variable and assigned value have different types"},
		{"let foo = 123\nlet bar\nfoo = bar", "cannot set variable to undefined"},
		{"foo = 123", "assignment to undeclared variable"},
		{"0 = 1", "invalid assignment target"},
		{"{", "expression expected"},
		{"{\n", "missing matching '}'"},
		{"{\n    let a\n    let a\n}", "variable with same name exists in current scope"},
		{"{\n    let a = a\n}", "cannot use local variable in own initialiser"},
		{"let a = true\n{\n    let a = a\n}", "cannot use local variable in own initialiser"},
		{"if true", "missing '{' after if condition"},
		{"if (true)\n    print(0)", "missing '{' after if condition"},
		{"if true {\n} else", "expected if or '{' after else"},
		{"else", "expression expected"},
		{"if 1 {\n}", "condition expression is not a Boolean"},
		{"0 and false", "condition expression is not a Boolean"},
		{"0 or false", "condition expression is not a Boolean"},
		{"\"\" and false", "condition expression is not a Boolean"},
		{"let a\na and false", "condition expression is not a Boolean"},
		{"while true", "missing '{' after while condition"},
		{"while (true)\n    print(0)", "missing '{' after while condition"},
		{"while 1 {\n}", "condition expression is not a Boolean"},
		{"for let i = 0", "';' expected"},
		{"for let i = 0; i < 5", "';' expected"},
		{"for let i = 0; i < 5; i = i + 1", "'{' expected"},
		{"function", "expected function name"},
		{"function foo", "expected '(' after function name"},
		{"function foo(", "expected ')' after function parameters"},
		{"function foo(bar", "expected ')' after function parameters"},
		{"function foo(bar, baz", "expected ')' after function parameters"},
		{"function foo()", "expected '{' before function body"},
		{"function foo(bar)", "expected '{' before function body"},
		{"function foo(bar, baz)", "expected '{' before function body"},
		{"0()", "expression is not callable"},
		{"false()", "expression is not callable"},
		{"function foo(bar) {\n    let bar\n}", "variable with same name exists in current scope"},
		{"function foo(bar) {\n}\nfoo()", "expected 1 arguments, got 0"},
		{"function foo() {\n}\nfoo(0)", "expected 0 arguments, got 1"},
		{"return 0 0", "newline or eof expected"},
		{"{ 0", "missing matching '}'"},
		{"(function", "expected '('"},
		{"for ;; { 0 } 0", "newline or eof expected"},
		{"function foo() { 0 } 0", "newline or eof expected"},
		{"if false { 0 } 0", "newline or eof expected"},
		{"while false { 0 } 0", "newline or eof expected"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			ErrorTestRunner(t, tt.src, tt.want)
		})
	}
}

func TestErrorRegressions(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"if:", "unknown token"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			ErrorTestRunner(t, tt.src, tt.want)
		})
	}
}

func TestBackslash(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"\n", ""},
		{"\n\n", ""},
		{"1 \\\n+ 2", ""},
		{"print(1 \\\n+ 2)", "3"},
		{"print(1 \\\n+ 2 \\\n)", "3"},
		{"1 + \\\n2", ""},
		{"print(1 + \\\n2)", "3"},
		{"print(1 + \\\n2 \\\n)", "3"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			RegressionTestRunner(t, tt.src, tt.want)
		})
	}
}

func TestOps(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"print(1 + 2)", "3"},
		{`print("a" + "bc")`, `abc`},
		{"print(1 - 2)", "-1"},
		{"print(1 * 2)", "2"},
		{"print(1 / 2)", "0.5"},
		{"print(1 * (2 + 3) / (4 - 5))", "-5"},
		{"print(-(1 + 2))", "-3"},
		{"print(2 ** 2 ** 2 ** 2)", "65536"},
		{"print(4 ** 3 ** 2)", "262144"},
		{"print(!true)", "false"},
		{"print(!false)", "true"},
		{"print(1 == 1)", "true"},
		{"print(1 == 2)", "false"},
		{"print(0 == false)", "false"},
		{"print(1 == true)", "false"},
		{"print(1 != 1)", "false"},
		{"print(1 != 2)", "true"},
		{"print(0 != false)", "true"},
		{"print(1 != true)", "true"},
		{"print(1 < 2)", "true"},
		{"print(1 <= 2)", "true"},
		{"print(1 > 2)", "false"},
		{"print(1 >= 2)", "false"},
		{"print(1 <= 1)", "true"},
		{"print(1 >= 1)", "true"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			RegressionTestRunner(t, tt.src, tt.want)
		})
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{`print("")`, ``},
		{`print("\a\b\f\n\r\t\v\\\"")`, "\a\b\f\n\r\t\v\\\""},
		{`print("\x7f\u00A0\U0010ffff")`, "\x7f\u00a0\U0010ffff"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			RegressionTestRunner(t, tt.src, tt.want)
		})
	}
}

func TestPanicMode(t *testing.T) {
	s := "1 + \n\"hello"
	wants := []string{"expression expected", "unterminated string"}

	v := newTestVm()
	got, errs := v.Run(s, "vm_test.go")
	if errs == nil {
		t.Errorf("No error occurred, got %q, want %d errors", got, len(wants))
	} else {
		if len(errs) != len(wants) {
			t.Errorf("got %d errors, want %d errors", len(errs), len(wants))
			return
		}

		for i, err := range errs {
			if errors.Unwrap(err).Error() != wants[i] {
				t.Errorf("error #%d: got error %q, want error %q", i+1, errors.Unwrap(err), wants[i])
			}
		}
	}
}

func TestVariables(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"let foo = 123\nprint(foo)", "123"},
		{"let foo\nprint(foo)", "undefined"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			RegressionTestRunner(t, tt.src, tt.want)
		})
	}
}

func TestShadowing(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"let foo = 123\n{\n    let foo = true\n    print(foo)\n}", "true"},
		{"let foo = 123\n{\n    let foo = false\n}\nprint(foo)", "123"},
		{"let foo = 123\n{\n    let foo = false\n    {\n        let foo = \"abc\"\n        print(foo)\n    }\n}", "abc"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			RegressionTestRunner(t, tt.src, tt.want)
		})
	}
}

func TestControlFlow(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"if true {\n    print(\"true\")\n}", "true"},
		{"if false {\n    print(\"false\")\n}", ""},
		{"if true {\n    print(\"true\")\n} else {\n    print(\"false\")\n}", "true"},
		{"if false {\n    print(\"true\")\n} else {\n    print(\"false\")\n}", "false"},
		{"if false {\n    print(\"true\")\n} else if false {\n    print(\"false\")\n} else {\n    print(\"hi\")\n}", "hi"},
		{"print(true and true)", "true"},
		{"print(true and false)", "false"},
		{"print(false and true)", "false"},
		{"print(false and false)", "false"},
		{"print(true or true)", "true"},
		{"print(true or false)", "true"},
		{"print(false or true)", "true"},
		{"print(false or false)", "false"},
		{"let i = 1\nwhile i <= 5 {\n    print(i)\n    i = i + 1\n}", "1\n2\n3\n4\n5"},
		{"for let i = 5; i >= 1; i = i - 1 {\n    print(i)\n}", "5\n4\n3\n2\n1"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			RegressionTestRunner(t, tt.src, tt.want)
		})
	}
}

func TestFunctions(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"function foo() {\n}\nprint(foo())", "undefined"},
		{"function foo() {\n    print(foo)\n}\nfoo()", "<function foo>"},
		{"function foo() {\n    print(\"foo\")\n}\nfoo()", "foo"},
		{"function foo() {\n    print(\"foo\")\n}\nprint(foo())", "foo\nundefined"},
		{"function foo() {\n    return \"foo\"\n}\nfunction bar() {\n    return foo\n}\nprint(bar)", "<function bar>"},
		{"function foo() {\n    return \"foo\"\n}\nfunction bar() {\n    return foo\n}\nprint(bar())", "<function foo>"},
		{"function foo() {\n    return \"foo\"\n}\nfunction bar() {\n    return foo\n}\nprint(bar()())", "foo"},
		{"return", ""},
		{"return 0", ""},
		{"print(function() { return 123 }())", "123"},
		{"(function() { print(123) })()", "123"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			RegressionTestRunner(t, tt.src, tt.want)
		})
	}
}

func TestCompoundAssign(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"let i = 1\ni += 2\nprint(i)", "3"},
		{"let i = 1\ni -= 2\nprint(i)", "-1"},
		{"let i = 1\ni *= 2\nprint(i)", "2"},
		{"let i = 1\ni /= 2\nprint(i)", "0.5"},
		{"let i = 1\ni **= 2\nprint(i)", "1"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			RegressionTestRunner(t, tt.src, tt.want)
		})
	}
}
