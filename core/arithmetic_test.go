package core

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	curve := &EllipticCurve{Name: "y^2=x^3+3x+8"}
	curve.A = big.NewInt(3)
	curve.B = big.NewInt(8)
	curve.P = big.NewInt(13)
	p1 := &Point{big.NewInt(9), big.NewInt(7)}
	p2 := &Point{big.NewInt(1), big.NewInt(8)}
	zero := &Point{big.NewInt(0), big.NewInt(0)}
	// test 1: P1 + P2
	p3 := curve.Add(p1, p2)
	assert.True(t, curve.Equal(p3, &Point{big.NewInt(2), big.NewInt(10)}))
	// test 2: P1 + P1
	p3 = curve.Add(p1, p1)
	assert.True(t, curve.Equal(p3, &Point{big.NewInt(9), big.NewInt(6)}))
	// test 3: P2 + P2
	p3 = curve.Add(p2, p2)
	assert.True(t, curve.Equal(p3, &Point{big.NewInt(2), big.NewInt(3)}))
	// test 4: P1 + 0 = P1
	p3 = curve.Add(p1, zero)
	assert.True(t, curve.Equal(p3, p1))
	assert.False(t, curve.Equal(p3, p2))
	assert.False(t, curve.Equal(p3, zero))
	// test 5: 0 + P2 = P2
	p3 = curve.Add(zero, p2)
	assert.True(t, curve.Equal(p3, p2))
	assert.False(t, curve.Equal(p3, zero))
	// test 6: P1-P2
	p3 = curve.Add(p1, curve.Negate(p2))
	assert.True(t, curve.Equal(p3, &Point{big.NewInt(12), big.NewInt(2)}))
}

/*
Check that addition works using industry param. Points and results are from golang's p256 curve.

n is most likely used to generate private key. The size of the private key is similar to size of the order of the base point, which is similar to curve's BitSize. For p256, it's 256bits. `make` allocate bytes so that's 32 bytes.

We're using p1 and p2 as points on the curve. Since we don't want to test multiplication before adding, we're getting the points from golang curve p256 by multiplying the base point with two random numbers. So p1=n1*G, p2=n2*G.
*/
func TestAddLarge(t *testing.T) {
	// set up golang's curve and points
	p256 := elliptic.P256()
	fmt.Println("golang curve", p256.Params().Name)
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	p1X, p1Y := p256.ScalarBaseMult(b)
	fmt.Println("p1", p1X, p1Y)
	_, _ = rand.Read(b)
	p2X, p2Y := p256.ScalarBaseMult(b)
	fmt.Println("p2", p2X, p2Y)
	p3X, p3Y := p256.Add(p1X, p1Y, p2X, p2Y)
	fmt.Println("p3=p1+p2", p3X, p3Y)

	// setup our curve
	curve := P256()
	p1 := &Point{p1X, p1Y}
	p2 := &Point{p2X, p2Y}
	p3 := curve.Add(p1, p2)
	assert.True(t, curve.Equal(p3, &Point{p3X, p3Y}))
}

func TestScalarMult(t *testing.T) {
	// Hoffstein example 6.16, n=947 Multiply nP
	curve := &EllipticCurve{Name: "y^2=x^3+14x+19"}
	curve.A = big.NewInt(14)
	curve.B = big.NewInt(19)
	curve.P = big.NewInt(3623)
	// n should usually be in []bytes but we're testing a particular value here
	n := big.NewInt(947)
	p := &Point{X: big.NewInt(6), Y: big.NewInt(730)}
	np := curve.ScalarMult(n.Bytes(), p)
	assert.True(t, curve.Equal(np, &Point{big.NewInt(3492), big.NewInt(60)}))
}

func TestScalarMultLarge(t *testing.T) {
	// set up golang's curve and points
	p256 := elliptic.P256()
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	pX, pY := p256.ScalarBaseMult(b)
	_, _ = rand.Read(b)
	npX, npY := p256.ScalarMult(pX, pY, b)

	// setup our curve
	curve := P256()
	p := &Point{pX, pY}
	np := curve.ScalarMult(b, p)
	assert.True(t, curve.Equal(np, &Point{npX, npY}))
}

func TestPreprocessTrits(t *testing.T) {
	n, _ := strconv.ParseInt("100110111001", 2, 32) // n=2489
	trits := preprocessTrits(big.NewInt(n).Bytes())
	assert.Equal(t, trits, []int8{1, 0, 0, -1, 0, 0, -1, 0, 0, 1, 0, 1, 0, 0, 0, 0})

	n, _ = strconv.ParseInt("10011110111001", 2, 32) // n=10169
	trits = preprocessTrits(big.NewInt(n).Bytes())
	assert.Equal(t, trits, []int8{1, 0, 0, -1, 0, 0, -1, 0, 0, 0, 0, 1, 0, 1, 0, 0})
}

func TestScalarMultTernary(t *testing.T) {
	curve := &EllipticCurve{Name: "y^2=x^3+14x+19"}
	curve.A = big.NewInt(14)
	curve.B = big.NewInt(19)
	curve.P = big.NewInt(3623)
	p := &Point{X: big.NewInt(6), Y: big.NewInt(730)}
	n, _ := strconv.ParseInt("100110111001", 2, 32) // 2489
	np := curve.ScalarMult(big.NewInt(2489).Bytes(), p)
	npTernary := curve.ScalarMultTernary(big.NewInt(n).Bytes(), p)
	assert.True(t, curve.Equal(np, &Point{big.NewInt(3241), big.NewInt(2032)}))
	assert.True(t, curve.Equal(npTernary, &Point{big.NewInt(3241), big.NewInt(2032)}))
}

func TestScalarMultTernaryLarge(t *testing.T) {
	// set up golang's curve and points
	p256 := elliptic.P256()
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	pX, pY := p256.ScalarBaseMult(b)
	_, _ = rand.Read(b)
	npX, npY := p256.ScalarMult(pX, pY, b)

	// setup our curve
	curve := P256()
	p := &Point{pX, pY}
	np := curve.ScalarMult(b, p)
	assert.True(t, curve.Equal(np, &Point{npX, npY}))
	npTernary := curve.ScalarMultTernary(b, p)
	assert.True(t, curve.Equal(npTernary, &Point{npX, npY}))
}

func TestIsOnCurve(t *testing.T) {
	curve := &EllipticCurve{Name: "y^2=x^3+14x+19"}
	curve.A = big.NewInt(14)
	curve.B = big.NewInt(19)
	curve.P = big.NewInt(3623)

	p := &Point{X: big.NewInt(3241), Y: big.NewInt(2032)}
	assert.True(t, curve.IsOnCurve(p))

	p = &Point{X: big.NewInt(3241), Y: big.NewInt(2031)}
	assert.False(t, curve.IsOnCurve(p))
}
