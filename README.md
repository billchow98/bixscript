## bixscript
A bytecode interpreter written in Go for a simple scripting language.

### Features
- Stack-based bytecode virtual machine runtime
- Recursive descent parser
- Value types
  - `Number` (double-precision floating-point type)
  - `Boolean` (`true` or `false`)
  - `String`
  - `undefined`
  - `function`
- Operators
  - Arithmetic: `+`, `-`, `*`, `/`, `**`
  - Compound assignment: `+=`, `-=`, `*=`, `/=`, `**=`
  - Comparison: `<`, `<=`, `==`, `!=`, `>`, `>=`
  - Logical: `and`, `or`
- Line comments
- Local and global variables
- `if-else` statements
- `for` and `while` loops
- Functions
- Support for registering native Go functions at compile-time, including
variadic functions
- Automatic garbage collection using Go's garbage collector
- Pretty-printed bytecode assembly for debugging purposes
- Comprehensive unit tests for every feature

### Further extensions
- Closures
- User-defined classes
- Standard library

### Examples
#### Naive Fibonacci
Code
```javascript
function fib(n) {
    if n <= 1 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

print(fib(35))
```

Sample REPL output with bytecode assembly
```javascript
2024/09/14 12:16:48 bix.go:20: bix v0.0.1
>>> function fib(n) {
...     if n <= 1 {
...         return n
...     }
...     return fib(n - 1) + fib(n - 2)
... }
Disassembly (function fib)
00      12 7c           Ldar r(-4)
02      11 80           Star r0
04      00 00           LdaConstant [0]
06      0b 00           CmpLe r0
08      15 83           JumpIfFalse [5], (13)
10      12 7c           Ldar r(-4)
12      17              Return
13      10 01           LdaGlobal [1]
15      11 80           Star r0
17      12 7c           Ldar r(-4)
19      11 81           Star r1
21      00 02           LdaConstant [2]
23      02 01           Sub r1
25      11 81           Star r1
27      16 05 01        Call [5], [1]
30      11 80           Star r0
32      10 01           LdaGlobal [1]
34      11 81           Star r1
36      12 7c           Ldar r(-4)
38      11 82           Star r2
40      00 03           LdaConstant [3]
42      02 02           Sub r2
44      11 82           Star r2
46      16 06 01        Call [6], [1]
49      01 00           Add r0
51      17              Return
52      18              LdaUndefined
53      17              Return
Constant pool (size 4)
0       <Number>        1
1       <Identifier>    fib
2       <Number>        1
3       <Number>        2
Disassembly (function <module>)
0       0e 00           GlobDecl [0]
2       00 01           LdaConstant [1]
4       0f 00           StaGlobal [0]
6       18              LdaUndefined
7       17              Return
Constant pool (size 2)
0       <Identifier>    fib
1       <Function>      <function fib>
>>> print(fib(35))
Disassembly (function <module>)
00      10 00           LdaGlobal [0]
02      11 80           Star r0
04      10 01           LdaGlobal [1]
06      11 81           Star r1
08      00 02           LdaConstant [2]
10      11 82           Star r2
12      16 06 01        Call [6], [1]
15      11 81           Star r1
17      16 05 01        Call [5], [1]
20      18              LdaUndefined
21      17              Return
Constant pool (size 3)
0       <Identifier>    print
1       <Identifier>    fib
2       <Number>        35
9227465
>>>
```

#### [too simple](https://benchmarksgame-team.pages.debian.net/benchmarksgame/performance/toosimple.html) micro-benchmark
- 10 million instead of 10 billion iterations

Code
```javascript
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
```

Sample REPL output with bytecode assembly
```javascript
2024/09/14 12:38:35 bix.go:20: bix v0.0.1
>>> function foo() {
...     let N = 10000000
...     let sum = 0.0
...     let flip = -1.0
...     for let i = 1; i <= N; i += 1 {
...         flip *= -1.0
...         sum += flip / (2*i - 1)
...     }
...     print(fmtSprintf("%.9f", sum*4.0))
... }
Disassembly (function foo)
000     00 00           LdaConstant [0]
002     11 80           Star r0
004     00 01           LdaConstant [1]
006     11 81           Star r1
008     00 02           LdaConstant [2]
010     11 82           Star r2
012     00 03           LdaConstant [3]
014     11 83           Star r3
016     12 83           Ldar r3
018     11 84           Star r4
020     12 80           Ldar r0
022     0b 04           CmpLe r4
024     15 b2           JumpIfFalse [52], (76)
026     12 82           Ldar r2
028     11 84           Star r4
030     00 04           LdaConstant [4]
032     03 04           Mul r4
034     11 82           Star r2
036     12 81           Ldar r1
038     11 84           Star r4
040     12 82           Ldar r2
042     11 85           Star r5
044     00 05           LdaConstant [5]
046     11 86           Star r6
048     12 83           Ldar r3
050     03 06           Mul r6
052     11 86           Star r6
054     00 06           LdaConstant [6]
056     02 06           Sub r6
058     04 05           Div r5
060     01 04           Add r4
062     11 81           Star r1
064     12 83           Ldar r3
066     11 84           Star r4
068     00 07           LdaConstant [7]
070     01 04           Add r4
072     11 83           Star r3
074     13 44           Jump [-58], (16)
076     10 08           LdaGlobal [8]
078     11 83           Star r3
080     10 09           LdaGlobal [9]
082     11 84           Star r4
084     00 0a           LdaConstant [10]
086     11 85           Star r5
088     12 81           Ldar r1
090     11 86           Star r6
092     00 0b           LdaConstant [11]
094     03 06           Mul r6
096     11 86           Star r6
098     16 0a 02        Call [10], [2]
101     11 84           Star r4
103     16 08 01        Call [8], [1]
106     18              LdaUndefined
107     17              Return
Constant pool (size 12)
00      <Number>        10000000
01      <Number>        0
02      <Number>        -1
03      <Number>        1
04      <Number>        -1
05      <Number>        2
06      <Number>        1
07      <Number>        1
08      <Identifier>    print
09      <Identifier>    fmtSprintf
10      <String>        "%.9f"
11      <Number>        4
Disassembly (function <module>)
0       0e 00           GlobDecl [0]
2       00 01           LdaConstant [1]
4       0f 00           StaGlobal [0]
6       18              LdaUndefined
7       17              Return
Constant pool (size 2)
0       <Identifier>    foo
1       <Function>      <function foo>
>>> foo()
Disassembly (function <module>)
0       10 00           LdaGlobal [0]
2       11 80           Star r0
4       16 04 00        Call [4], [0]
7       18              LdaUndefined
8       17              Return
Constant pool (size 1)
0       <Identifier>    foo
3.141592554
>>>
```

### Credits
This project was inspired by the book
[*Crafting Interpreters*](https://craftinginterpreters.com/) by Robert Nystrom.

### License
Copyright © 2023 Bill Chow. All rights reserved.

This repository is made available for evaluation purposes by authorized individuals only.
Unauthorized use, reproduction, modification, or distribution of this code is strictly prohibited.