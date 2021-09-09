package core

import (
	"crypto/elliptic"
	"crypto/rand"
)

/* KeyPair represents a public/private key pair for ECC
The private key is a random number size equal to N, the order of the base point
The public key is a point on the curve that's a multiple of the private key
and the base point.
*/
type KeyPair struct {
	Priv []byte
	Pub  *Point
}

func GenerateKey(curve elliptic.Curve) (*KeyPair, error) {
	// size of the private key is similar to size of the order of the base point
	priv := make([]byte, curve.Params().N.BitLen()/8)
	_, err := rand.Read(priv)
	if err != nil {
		return nil, err
	}
	point := &Point{nil, nil}
	point.X, point.Y = curve.ScalarBaseMult(priv)
	pair := &KeyPair{priv, point}
	return pair, nil
}
