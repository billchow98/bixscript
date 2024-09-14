// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package lib

// Contains aliases to actual library functions for convenience
func init() {
	registerNative("print", VarArgc, fmtPrintln)
}
