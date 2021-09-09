package core

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
)

/*
TODO:
1. Make sure points are on curve before operations, otherwise vulnerable to attack.
2. Both ScalarMult and ScalarMultTernary is vulnerable to timing attack. Implement constant time mult.
3. Implement more efficient way to convert to NAF (Non-Adjacent Form).
*/

/*
Point represents a point on the elliptic curve. It uses pointer to big.Int instead of the value itself because the API for big.Int is mostly built on pointers.

The point at infinity \mathcal{0} is represented by (0,0) even though the point (0,0) itself might not be on the curve. This is because I lack a way to represent it nicely so that I can do comparison and allocation. Note that the Add method will return P if given (0,0)+P or P+(0,0).
*/
type Point struct {
	X *big.Int
	Y *big.Int
}

/*
EllipticCurve represents an elliptic curve of the form y^2 = x^3+Ax+B x,y in F_p.
It differs from the golang ECC lib's CurveParams in that we can specify both of the constants
A and B.
*/
type EllipticCurve struct {
	P       *big.Int // order of the underlying field
	N       *big.Int // order of the base point
	A, B    *big.Int // constants of the curve equation
	G       *Point   // base point
	BitSize int      // size of the underlying field in bits
	Name    string   // name of the curve
}

/* Equal returns true if two points are equal on the curve.
 */
func (curve *EllipticCurve) Equal(p1, p2 *Point) bool {
	if p1.X.Cmp(p2.X) == 0 && p1.Y.Cmp(p2.Y) == 0 {
		return true
	} else {
		return false
	}
}

/* EqualPointAtInfinity returns true if the point is equal to the point at infinity, irrespective of what we finally choose to be the representation of point at infinity. For now, it's (0,0).
 */
func (curve *EllipticCurve) EqualPointAtInfinity(p *Point) bool {
	// to test if a big.Int is zero, it's apparently faster to
	// check the length of its bits
	// https://stackoverflow.com/questions/64257065/is-there-another-way-of-testing-if-a-big-int-is-0
	if len(p.X.Bits()) == 0 && len(p.Y.Bits()) == 0 {
		return true
	} else {
		return false
	}
}

/*
IsOnCurve returns true if the point is on the curve.
Calculate x^3-Ax+B and compare it with y^2.
It will return true if p is (0,0), even if the point is not technically on the curve.
This is useful for detecting illegal curve operations where the point is not even on the curve.
*/
func (curve *EllipticCurve) IsOnCurve(p *Point) bool {
	x3 := new(big.Int).Mul(p.X, p.X)
	x3.Mul(x3, p.X)
	xA := new(big.Int).Set(curve.A)
	xA.Mul(xA, p.X)
	x3.Add(x3, xA)
	x3.Add(x3, curve.B)
	x3.Mod(x3, curve.P)

	y2 := new(big.Int).Mul(p.Y, p.Y)
	y2.Mod(y2, curve.P)
	return y2.Cmp(x3) == 0
}

/*
Add two points on the curve, return a new point.
This uses the very naive approach of calculating the slope of the line going
through the two points.
If either of the point is (0,0), return the other point.
NOTE: the points passed to Add are supposed to be on the curve, this will still work even if one of the points is not on the curve. Checking has to be done elsewhere.
TODO: handle cases: (1) both is 0 (2)they add up to 0.
*/
func (curve *EllipticCurve) Add(p1, p2 *Point) *Point {
	// if p1=(0,0), return a copy of p2
	if curve.EqualPointAtInfinity(p1) {
		return &Point{new(big.Int).Set(p2.X), new(big.Int).Set(p2.Y)}
	}
	// if p2=(0,0), return a copy of p1
	if curve.EqualPointAtInfinity(p2) {
		return &Point{new(big.Int).Set(p1.X), new(big.Int).Set(p1.Y)}
	}
	// otherwise neither of the point is (0,0)
	lambda := new(big.Int)
	top := new(big.Int)
	bottom := new(big.Int)
	if p1.X.Cmp(p2.X) != 0 && p1.Y.Cmp(p2.Y) != 0 {
		top.Sub(p2.Y, p1.Y)
		bottom.Sub(p2.X, p1.X)
		bottom.ModInverse(bottom, curve.P)
		lambda.Mul(top, bottom)
		lambda.Mod(lambda, curve.P)
	} else if curve.Equal(p1, p2) {
		top.Mul(p1.X, p1.X)
		top.Mul(top, big.NewInt(3))
		top.Add(top, curve.A)
		top.Mod(top, curve.P)
		bottom.Mul(p1.Y, big.NewInt(2))
		bottom.ModInverse(bottom, curve.P)
		lambda.Mul(top, bottom)
		lambda.Mod(lambda, curve.P)
	}
	p := &Point{new(big.Int), new(big.Int)}
	p.X.Mul(lambda, lambda)
	p.X.Sub(p.X, p1.X)
	p.X.Sub(p.X, p2.X)
	p.X.Mod(p.X, curve.P)
	p.Y.Sub(p1.X, p.X)
	p.Y.Mul(lambda, p.Y)
	p.Y.Sub(p.Y, p1.Y)
	p.Y.Mod(p.Y, curve.P)
	return p
}

/*
Negate the point on the curve, i.e. return a point with -y. This should not change the point p.
This function exists because negation of big.Int is fiddly and change its value.
*/
func (curve *EllipticCurve) Negate(p *Point) *Point {
	// make a copy of p so that we don't change it
	p1 := &Point{new(big.Int).Set(p.X), new(big.Int).Set(p.Y)}
	p1.Y.Neg(p.Y)
	return p1
}

/*
Use the double and add method to calculate nP. Essentially, write n in binary
form, doubling P as we go along the bitstring of n, adding the result
when we see a 1 bit. See Hoffstein 6.3.1.

Note that this is vulnerable to timing analysis:
https://en.wikipedia.org/wiki/Elliptic_curve_point_multiplication

Generally, the size of n is going to be the order of the base point, so we use []byte for n. It's likely to be randomly generated by reading from rand.
*/
func (curve *EllipticCurve) ScalarMult(n []byte, p *Point) *Point {
	// converting n into big.Int because we need to right shift
	// TODO (later) write function to shift byte slice so that
	// we don't have to convert n to big.Int
	scalar := new(big.Int).SetBytes(n)
	ret := &Point{big.NewInt(0), big.NewInt(0)}
	doubles := &Point{new(big.Int).Set(p.X), new(big.Int).Set(p.Y)}
	additions := 0
	// carry on computation until n is 0
	for len(scalar.Bits()) != 0 {
		if scalar.Bit(0) == 1 { // time to add things up
			ret = curve.Add(ret, doubles)
			additions++
		}
		doubles = curve.Add(doubles, doubles)
		scalar.Rsh(scalar, 1)
	}
	// the first addtion P+(0,0) doesn't count
	fmt.Println("Total number of additions:", additions-1)
	return ret
}

/*
ScalarMultTernary uses the ternary expansion of n.
First preprocess the bits of n to get rid of contiguous groups of 1s with size > 2.
Then use double and add/subtract to calculate nP.
*/
func (curve *EllipticCurve) ScalarMultTernary(n []byte, p *Point) *Point {
	trits := preprocessTrits(n)
	ret := &Point{big.NewInt(0), big.NewInt(0)}
	doubles := &Point{new(big.Int).Set(p.X), new(big.Int).Set(p.Y)}
	additions := 0
	for i := 0; i < len(trits); i++ {
		if trits[i] == 1 {
			ret = curve.Add(ret, doubles)
			additions++
		}
		if trits[i] == -1 {
			ret = curve.Add(ret, curve.Negate(doubles))
			additions++
		}
		doubles = curve.Add(doubles, doubles)
	}
	// the first addtion P+(0,0) doesn't count
	fmt.Println("Total number of additions:", additions-1)
	return ret
}

/*
preprocessTrits turns n into a ternary array, using 0, 1 and -1. The result should have no contiguous group of 1s with length>2 ("bad group").
*/
func preprocessTrits(n []byte) []int8 {
	// unpack n into bits so that the lowest bit is to the left
	trits := make([]int8, len(n)*8)
	bitLen := len(n)*8 - 1
	for i, b := range n {
		for j := 0; j < 8; j++ {
			trits[bitLen-(i*8+j)] = int8(b >> uint(7-j) & 1)
		}
	}
	// process bits to get rid of bad groups, s marks the start of a group
	// t marks the number of 1s in the group
	s, t := 0, 0
	for i := 0; i < len(trits); i++ {
		if trits[i] == 1 {
			if i >= 1 && trits[i-1] == 0 { // prev trit is 0
				s = i // mark the start of the group
			}
			t++
		} else if trits[i] == 0 {
			if t > 2 { // at the end of a bad group
				// turn start of the group into -1
				trits[s] = -1
				// flip the rest of the group to 0
				for j := s + 1; j < s+t; j++ {
					trits[j] = 0
				}
				// flip current bit (should be 0) to 1
				trits[i] = 1
				// mark the start of a new group
				s = i
				t = 1
			} else { // otherwise it wasn't a bad group, ignore
				t = 0 // reset group length
			}
		}
	}
	return trits
}

func Example() {
	// setup small curve, and point and n
	fmt.Println("***Testing with small example***")
	curve := &EllipticCurve{Name: "y^2=x^3+14x+19"}
	curve.A = big.NewInt(14)
	curve.B = big.NewInt(19)
	curve.P = big.NewInt(3623)
	fmt.Println(curve)
	p := &Point{X: big.NewInt(6), Y: big.NewInt(730)}
	n, _ := strconv.ParseInt("100110111001", 2, 32) // 2489

	/////////// Small example: Scalar mult with double and add
	np := curve.ScalarMult(big.NewInt(n).Bytes(), p)
	fmt.Println(np, "ScalarMult with Double and Add")

	/////////// Small example: Scalar mult with ternary expansion
	npTernary := curve.ScalarMultTernary(big.NewInt(n).Bytes(), p)
	fmt.Println(npTernary, "ScarlarMult with Ternary Expansion")

	// setup industrial size curve p256, and point and n
	// set up golang's curve and points
	fmt.Println("***Testing with P-256***")
	p256 := elliptic.P256()
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	pX, pY := p256.ScalarBaseMult(b)
	_, _ = rand.Read(b)
	npX, npY := p256.ScalarMult(pX, pY, b)

	// setup our curve
	curve = P256()
	p = &Point{pX, pY}
	fmt.Println("ScalarMult with Double and Add")
	np = curve.ScalarMult(b, p)
	fmt.Println("ScalarMult correct?", np.X.Cmp(npX) == 0 && np.Y.Cmp(npY) == 0)
	fmt.Println("ScalarMult with Ternary Expansion")
	npTernary = curve.ScalarMultTernary(b, p)
	fmt.Println("ScalarMultTernary correct?", np.X.Cmp(npX) == 0 && np.Y.Cmp(npY) == 0)
}

/*
Returns a curve that uses golang p256 curve parameters, which is based on FIPS 186-3, section D.2.3 (p.100). Per this guide, section D.1.2 specifies the curve equation:
$$E:y^2 \equiv x^3-3x+b (mod p)$$

Golang's implementation:
https://cs.opensource.google/go/go/+/refs/tags/go1.16.6:src/crypto/elliptic/p256.go

NIST publication:
https://csrc.nist.gov/csrc/media/publications/fips/186/3/archive/2009-06-25/documents/fips_186-3.pdf
*/
func P256() *EllipticCurve {
	curve := &EllipticCurve{Name: "P-256"}
	curve.A = big.NewInt(-3)
	curve.B, _ = new(big.Int).SetString("5ac635d8aa3a93e7b3ebbd55769886bc651d06b0cc53b0f63bce3c3e27d2604b", 16)
	curve.P, _ = new(big.Int).SetString("115792089210356248762697446949407573530086143415290314195533631308867097853951", 10)
	curve.N, _ = new(big.Int).SetString("115792089210356248762697446949407573529996955224135760342422259061068512044369", 10)
	p := &Point{}
	p.X, _ = new(big.Int).SetString("6b17d1f2e12c4247f8bce6e563a440f277037d812deb33a0f4a13945d898c296", 16)
	p.Y, _ = new(big.Int).SetString("4fe342e2fe1a7f9b8ee7eb4a7c0f9e162bce33576b315ececbb6406837bf51f5", 16)
	curve.G = p
	curve.BitSize = 256
	return curve
}
