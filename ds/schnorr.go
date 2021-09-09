package ds

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

/*
This is purposely done very naively using golang elliptic lib, without the
rest of this lib.
*/
func SchnorrExample() {
	fmt.Println("------ Schnoor example ------")
	// key generation
	curve := elliptic.P256()
	priv, x, y, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("private key size", len(priv), priv)
	fmt.Println("public key", x, y)

	// message
	msg := "hello, world"
	digest := sha256.Sum256([]byte(msg))
	fmt.Printf("message digest type %T, digest %v\n", digest, digest)

	// signing
	s, e, err := sign(curve, priv, digest[:])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("signature", s, e)

	// verifying
	fmt.Println("Schnoor works?", verify(curve, x, y, s, e, digest[:]))
}

func sign(curve elliptic.Curve, priv, digest []byte) (s, e *big.Int, err error) {
	N := curve.Params().N
	// k is a random elt of F_N, this is purposely not following NSA A.2.1
	b := make([]byte, curve.Params().BitSize/8)
	_, err = rand.Read(b)
	if err != nil {
		fmt.Println(err)
		return
	}
	k := new(big.Int).SetBytes(b)
	k.Mod(k, N)
	fmt.Println("k", k.BitLen(), k)

	r_x, _ := curve.ScalarBaseMult(k.Bytes())
	fmt.Println("x(r)", r_x.BitLen(), r_x)

	md := sha256.New()
	md.Write(r_x.Bytes())
	md.Write(digest[:])
	e = new(big.Int).SetBytes(md.Sum(nil))

	s = new(big.Int).SetBytes(priv)
	s.Mul(s, e)
	s.Sub(k, s)
	s.Mod(s, N)
	return s, e, nil
}

func verify(curve elliptic.Curve, x, y, s, e *big.Int, digest []byte) bool {
	tmp1_x, tmp1_y := curve.ScalarMult(x, y, e.Bytes())
	tmp2_x, tmp2_y := curve.ScalarBaseMult(s.Bytes())
	rv_x, _ := curve.Add(tmp1_x, tmp1_y, tmp2_x, tmp2_y)
	fmt.Println("rv_x", rv_x)

	md := sha256.New()
	md.Write(rv_x.Bytes())
	md.Write(digest)
	e_v := new(big.Int).SetBytes(md.Sum(nil))
	fmt.Println("e_v", e_v)
	return e_v.Cmp(e) == 0
}
