# © 2023 Bill Chow. All rights reserved.
# Unauthorized use, modification, or distribution of this code is strictly
# prohibited.

def foo():
    N = 10000000
    sum = 0.0
    flip = -1.0
    for i in range(1, N + 1):
        flip *= -1.0
        sum += flip / (2 * i - 1)
    print(sum * 4.0)

foo()
