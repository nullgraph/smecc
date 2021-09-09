package core

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
)

// Example using the golang built-in ECC lib
func GoECCExample() {
	curve := elliptic.P256()
	fmt.Println(curve.Params().Name)
	curveParams := curve.Params()
	fmt.Println("------ curve ------")
	fmt.Println("order of underlying field, P", curveParams.P, "bitsize", curveParams.P.BitLen())
	fmt.Println("order of base point, N", curveParams.N, "bitsize", curveParams.N.BitLen())
	fmt.Println("size of underlying field, BitSize", curveParams.BitSize)
	fmt.Println("B", curveParams.B)
	fmt.Println("base point", curveParams.Gx, curveParams.Gy)
	fmt.Println("name", curveParams.Name)

	// 20 is apparently a common choice for a false positive rate 0.000,000,000,001
	fmt.Println("order of base point N is prime?", curveParams.N.ProbablyPrime(20))

	// Generate key pair
	keyPair, err := GenerateKey(curve)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("------ key pair ------")
	fmt.Println("private key", keyPair.Priv)
	fmt.Println("public key", keyPair.Pub.X, keyPair.Pub.Y)

	// Generate random point on the curve by n*P
	b1 := make([]byte, 32)
	_, err = rand.Read(b1)
	if err != nil {
		fmt.Println(err)
		return
	}
	pX, pY := curve.ScalarBaseMult(b1)
	fmt.Println("random point", pX.Text(16), pY.Text(16))
}
