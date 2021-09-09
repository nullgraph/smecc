package dh

import (
	"crypto/elliptic"
	"ecc/core"
	"fmt"
)

/*
The shared secret is pubB * privA = pubA * privB
This works because pubA and pubB are multiples of the base point by
privA and privB corresp. The shared secret is a point on the curve.
This function is simply to simulate the exchange between A and B, and to show
that everything would work.
*/
func DHExample() {
	fmt.Println("------ Diffie-Hellman ------")
	curve := elliptic.P256()

	// A generates her private/public key pair
	pairA, err := core.GenerateKey(curve)
	if err != nil {
		fmt.Println(err)
		return
	}
	// B generates his private/public key pair
	pairB, err := core.GenerateKey(curve)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Transmit step: A sends pairA.Pub and B sends pairB.Pub to each other.

	// A calculates her shared key
	sharedA := &core.Point{nil, nil}
	sharedA.X, sharedA.Y = curve.ScalarMult(pairB.Pub.X, pairB.Pub.Y, pairA.Priv)
	// B calculates his shared key
	sharedB := &core.Point{nil, nil}
	sharedB.X, sharedB.Y = curve.ScalarMult(pairA.Pub.X, pairA.Pub.Y, pairB.Priv)
	// verify that the shared keys are equal
	ret := (sharedA.X.Cmp(sharedB.X) == 0) && (sharedA.Y.Cmp(sharedB.Y) == 0)
	fmt.Println("DH works?", ret)
}

/*
shortDH verifies that a shorter implementation of DH works, similar to dh func.

In shortDH, A and B only transmits the x-coord of their public key, each party
then calculate the y-coord by taking square root mod ?. The shared key is the
x-coord of the shared point, which is well-defined.

In practice, this is probably rarely used since it makes sense for A and B to
publish their public keys.

TODO: abandon this because golang's ECC lib doesn't have an easy way to
calculate y-coord of a point given x-coord, lacking a straight forward
way to access the curve.
*/
/*
func shortDH(curve elliptic.Curve) (bool, error) {
	// A generates her private/public key pair
	pairA, err := generateKey(curve)
	if err != nil {
		return false, err
	}
	// B generates his private/public key pair
	pairB, err := generateKey(curve)
	if err != nil {
		return false, err
	}

	// Transmit step: A transmits pairA.Pub.X and B transmits pairB.Pub.X

	// A calculates her shared key
	// first calculate B's public key y-coord
	// then calculate shared key just as standard DH, the shared key is x-coord


	// B calculates his shared key

	// verify that the shared keys matched
}
*/
