package misc

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
)

func PedersenCommitmentExample() {
	curve := elliptic.P256()
	// set up G= b1*P and H=b2*P where P is base point
	b1 := make([]byte, 32)
	_, err := rand.Read(b1)
	if err != nil {
		fmt.Println(err)
	}
	Gx, Gy := curve.ScalarBaseMult(b1)
	b2 := make([]byte, 32)
	_, err = rand.Read(b2)
	if err != nil {
		fmt.Println(err)
	}
	Hx, Hy := curve.ScalarBaseMult(b2)

	// message
	m := make([]byte, 32)
	_, err = rand.Read(m)
	if err != nil {
		fmt.Println(err)
	}

	r := make([]byte, 32)
	_, err = rand.Read(r)
	if err != nil {
		fmt.Println(err)
	}

	// commit
	Cx, Cy, _ := commit(curve, m, r, Gx, Gy, Hx, Hy)
	fmt.Println(Cx, Cy)
	// open
	fmt.Println(open(curve, m, r, Gx, Gy, Hx, Hy, Cx, Cy))
}

func commit(curve elliptic.Curve, m, r []byte, Gx, Gy, Hx, Hy *big.Int) (Cx, Cy *big.Int, err error) {
	tmp1x, tmp1y := curve.ScalarMult(Gx, Gy, m)
	tmp2x, tmp2y := curve.ScalarMult(Hx, Hy, r)
	Cx, Cy = curve.Add(tmp1x, tmp1y, tmp2x, tmp2y)
	return Cx, Cy, nil
}

func open(curve elliptic.Curve, m, r []byte, Gx, Gy, Hx, Hy, Cx, Cy *big.Int) bool {
	tmp1x, tmp1y := curve.ScalarMult(Gx, Gy, m)
	tmp2x, tmp2y := curve.ScalarMult(Hx, Hy, r)
	Dx, Dy := curve.Add(tmp1x, tmp1y, tmp2x, tmp2y)

	return Dx.Cmp(Cx) == 0 && Dy.Cmp(Cy) == 0
}
