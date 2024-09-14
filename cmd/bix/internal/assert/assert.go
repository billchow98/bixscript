// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package assert

import (
	"fmt"
	"runtime"
)

func Assert(b bool) {
	if b {
		return
	}
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	panic(fmt.Errorf("%s:%d %s(): logic error", frame.File, frame.Line, frame.Function))
}
