// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package lib

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/value"
	"time"
)

func init() {
	registerNative("timeNow", 0, timeNow)
	registerNative("timeSince", 1, timeSince)
	registerNative("durationString", 1, durationString)
}

func timeNow(int, *value.Value) value.Value {
	now := time.Now()
	return value.FromPointer(value.Pointer(&now))
}

func timeSince(_ int, argv *value.Value) value.Value {
	t := *(*time.Time)(getArg(argv, 0).AsPointer())
	elapsed := time.Since(t)
	return value.FromInteger(value.Integer(elapsed))
}

func durationString(_ int, argv *value.Value) value.Value {
	elapsed := getArg(argv, 0).AsInteger()
	return value.FromString(value.String(time.Duration(elapsed).String()))
}
