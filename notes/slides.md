% smecc (a small ECC library)
% Ha T. Lam
% August 24, 2021

# Why smecc?

- Learning purpose
- Golang ECC lib doesn't allow parameter changes
- E.g. curve $y^2=x^2+Ax+B$ has $A=-3$ per NIST

# Add Two Points

Given $P_1$ and $P_2$ on the curve, want to find $P_3 = P_1+P_2$.

- If $P_1 \neq P_2$, find slope $\lambda$ of line $L$ that goes through the two points, $\lambda = \frac{y_2-y_1}{x_2-x_1}$
- If $P_1 = P_2$, find slope of tangent line at the point, $\lambda = \frac{3x_1^2+A}{2y_1}$

Calculate intersection of line $L$ with the curve and reflect that point across the x-axis: $x_3=\lambda^2-x_1-x_2$ and $y_3= \lambda(x_1-x_3)-y_1$

# Scalar Multiplication

Given a point $P$ on the curve and an integer $n$, find $nP$.

Naive method: add $P$ to itself $n$ times.

Cost: expensive because $n$ could be large, 256 bits.

# Scalar Multiplication Using Double and Add

Example: $y^2=x^3+14x+19, \ p=3623, \ P=(6,730)$

Want to find $nP$ where $n=2489=(100110111001)_b$.

We can write $n$ as:
$$n = 2^0+2^3+2^4+2^5+2^7+2^8+2^{11}$$

The algorithm:

- Doubles the point $P$ and keeps doubling
- Adds the doubles whenever the $i^{th}$ bit of $n$ is 1

*Total cost:* 11 doublings and 6 additions.

# Scalar Multiplication Using Ternary Expansion

Example: $y^2=x^3+14x+19, \ p=3623, \ P=(6,730)$

Want to find $nP$ where $n=2489=(100110111001)_b$.

Instead of binary expansion of $n$, notice:
$$2^3+2^4+2^5 = 2^3(1+2+2^2) = 2^3(2^3-1) = 2^6-2^3$$

Substitute back:
$$n = 2^0-2^3+2^6+2^7+2^8+2^{11}$$

Continue substituing:
$$n = 2^0-2^3-2^6+2^9+2^{11}$$

*Total cost:* 11 doublings, 4 additions, since on an elliptic curve,
$$P_1-P_2 = (x_1,y_1)+(x_2,-y_2)$$

# Ternary Implementation
- $n$ comes as `[]byte` so it has to be unpacked to `[]int8`, we need -1, 0, 1 values
- For ease of computations, unpack $n$ so that the lowest bit is to the left
- Identify and substitute contiguous group of 1s with length $\geq 2$, "bad groups":
  - Use `s` to mark where the bad group starts
  - Use `t` to mark how many 1s are in the group
  - When encounter a 1-bit, check if this is a new group (the previous bit is 0), if it is, set `s=1`, but always increment `t`
  - When encounter a 0-bit, check if we're in a bad group (`t>2`), if we are then the group has come to an end, flip the appropriate bits
  - If not in a bad group, e.g. only saw 010, reset the group length `t`

# Ternary Implementation Cost

Main steps:

- Unpack and preprocess `n`, get array of `trits` (ternary digits)
- Use double and add/subtract to calcualte `nP`

Unpacking takes 1 pass and preprocess takes another pass. Typical `n` for industry parameter is 256 bits, not too bad for both time and storage.

Still have to do same amount of doublings.

How many additions can we save?

- After preprocessing, there are no contiguous groups of 1s with length $\geq 2$
- So at most $\frac{1}{2}$ of bits are non-zero (could be 1 or -1)
- Maximum number of additions is $\frac{1}{2}log(n)$

# Demonstration

Two types of tests: small test case above and industry parameter.

Industry-size test:

- Use P-256 curve $y^2 = x^3-3x+b$
- Parameters: $b$, $p$ and base point from golang source code
- Point $P$ is a random multiplication of the base point
- $n$ is random `[]byte`
- Comparison of double and add vs ternary expansion in terms of additions.

# Architectural Remarks
- Golang ECC lib has nice API, follow it
- Points on curve are struct of type `(*big.Int, *big.Int)`
- Point of infinity is denoted `(0,0)` even though it might not be on the curve
- `Add` function checks if either of the points are `(0,0)`, otherwise, addition will not be accurate.

# Wishlist
- Both `ScalarMult` by Double and Add and Ternary Expansion are vulnerable against timing attack. Implement fixed time multiplication (Montgomery).
- Need to check that points are on the curve before operations, otherwise vulnerable to attacks.
- `Add` function uses naive slope calculation, could be faster?
- Rename Ternary Expansion to NAF (Non-Adjacent Form) and improve performance.
