# Notes on writing this ECC lib

## Architecture

Following almost exactly how golang architecure [their ECC lib](https://golang.org/src/crypto/elliptic/elliptic.go).

There's a set of arithmetic operations that can be done on the curve. The curve is a struct with all the parameters. The set of operations is defined as an interface, each function has the curve as the receiver. Grammatically, this makes sense, to add two points on the curve, write `curve.Add(p1, p2)`.

## Arithmethic lib

Golang `big` library is not very nice to do modular arithmetic in, a search reveals [goff](https://hackmd.io/@zkteam/goff), apparently written by ConsenSys!
