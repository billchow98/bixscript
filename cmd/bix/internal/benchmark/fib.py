# © 2023 Bill Chow. All rights reserved.
# Unauthorized use, modification, or distribution of this code is strictly
# prohibited.

def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)

print(fib(35))

