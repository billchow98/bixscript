package lib

import (
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/value"
	"unsafe"
)

func init() {
	registerNative("fmtPrintln", VarArgc, fmtPrintln)
	registerNative("fmtSprintf", VarArgc, fmtSprintf)
}

func fmtPrintln(argc int, argv *value.Value) value.Value {
	v := make([]any, argc)
	for i, arg := range unsafe.Slice(argv, argc) {
		v[i] = arg
	}

	logger.Println(v...)

	return value.Undefined()
}

func fmtSprintf(argc int, argv *value.Value) value.Value {
	a := make([]any, argc)
	for i, arg := range unsafe.Slice(argv, argc) {
		switch {
		case arg.IsNumber():
			a[i] = arg.AsNumber()
		case arg.IsBoolean():
			a[i] = arg.AsBoolean()
		default:
			a[i] = arg
		}
	}

	format := string(getArg(argv, 0).AsString())
	a = a[1:]

	return value.FromString(value.String(fmt.Sprintf(format, a...)))
}
