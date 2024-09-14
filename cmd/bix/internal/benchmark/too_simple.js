// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

function foo() {
    console.time("foo")
    let N = 10000000
    let sum = 0.0
    let flip = -1.0
    for (let i = 1; i <= N; i++) {
        flip *= -1.0
        sum += flip / (2*i - 1)
    }
    console.timeEnd("foo")
    console.log(sum*4.0)
}

foo()
