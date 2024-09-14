// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package benchmark

import "testing"

const tooSimpleSrc = `
function foo() {
    let N = 10000000
    let sum = 0.0
    let flip = -1.0
    for let i = 1; i <= N; i += 1 {
        flip *= -1.0
        sum += flip / (2*i - 1)
    }
    print(fmtSprintf("%.9f", sum*4.0))
}

foo()
`

func BenchmarkTooSimple(b *testing.B) {
	benchmark(b, tooSimpleSrc, "fib_test.go")
}
