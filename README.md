# smecc

## Why

This library is written as I'm learning ECC and reading the Golang crypto library. I like the Golang crypto library but it doesn't allow for just any curve, in particular, there's no way to specify different parameters (`A != -3`) for the curve. This makes it difficult to experiment with different types of curves. 

Because of the way the library is structured, it also makes it hard to access the curve directly, for example, to calculate $y$ given $x$ for P-256.

## Architecture

This library follows almost exactly how golang architecure [their ECC lib](https://golang.org/src/crypto/elliptic/elliptic.go). I find their API to be clean and easy to work with.

### Point at infinity

I do not have a good way to denote the point at infinity since each point on the curve is of type `(*big.Int, *big.Int)` and by intention, `big.Int` does not have a maximum value.

Here the point at infinity is denoted `(0,0)` even though this point itself might or might not be on the curve. Because of this uncertainty, the `Add` function checks if either of the points are `(0,0)` and return the other point in that case, otherwise, addition will not be accurate.

This is one of the difficulties of representing curves in the affine plane instead of projective.

## Security

This library is written for learning and experimenting purpose, do NOT use this library in production!

Some of the security problems are:

- Both `ScalarMult` by Double and Add and Ternary Expansion are vulnerable against timing attack.
- There is no check that points are on the curve before operations, users are responsible for this.

## Todo

1. Implement fixed time multiplication (Montgomery), this might not be possible for all curves.
2. Rename Ternary Expansion to NAF (Non-Adjacent Form) and improve performance.

## References

An Introduction to Mathematical Cryptography, Hoffstein et al.

## Name

smecc stands for "small ecc" and is not related to Red Dwarf :-)
