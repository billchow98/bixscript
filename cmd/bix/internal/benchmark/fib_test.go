// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package benchmark

import (
	"testing"
)

const fibSrc = `
function fib(n) {
    if n <= 1 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

print(fib(35))
`

func BenchmarkFib(b *testing.B) {
	benchmark(b, fibSrc, "fib_test.go")
}
