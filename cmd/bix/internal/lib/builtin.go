package lib

// Contains aliases to actual library functions for convenience
func init() {
	registerNative("print", VarArgc, fmtPrintln)
}
