package main

import (
	"fmt"

	"go.dedis.ch/kyber/v3/sign/eddsa"
	"go.dedis.ch/kyber/v3/suites"
)

var ed25519 suites.Suite

func init() {
	ed25519 = suites.MustFind("Ed25519")
}

func GenerateKey() (privkey []byte, pubkey []byte, err error) {
	priv := ed25519.Scalar().Pick(ed25519.RandomStream())
	pub := ed25519.Point().Mul(priv, nil)

	privkey, err = priv.MarshalBinary()
	if err != nil {
		return
	}

	pubkey, err = pub.MarshalBinary()
	return
}

func ComputeSharedSecret(privkey []byte, pubkey []byte) (secBuf []byte, err error) {
	priv := ed25519.Scalar()
	pub := ed25519.Point()

	err = priv.UnmarshalBinary(privkey)
	if err != nil {
		return
	}

	err = pub.UnmarshalBinary(pubkey)
	if err != nil {
		return
	}

	secret := ed25519.Point().Mul(priv, pub)
	secBuf, err = secret.MarshalBinary()
	return
}

func Sign(privkey []byte, pubkey []byte, msg []byte) (signature []byte, err error) {
	ed := &eddsa.EdDSA{
		Public: ed25519.Point(),
		Secret: ed25519.Scalar(),
	}

	err = ed.Secret.UnmarshalBinary(privkey)
	if err != nil {
		return
	}

	err = ed.Public.UnmarshalBinary(pubkey)
	if err != nil {
		return
	}

	signature, err = ed.Sign(msg)
	return
}

func Verify(pubkey []byte, msg []byte, sig []byte) error {
	pub := ed25519.Point()

	err := pub.UnmarshalBinary(pubkey)
	if err != nil {
		return err
	}

	if err := eddsa.Verify(pub, msg, sig); err != nil {
		return fmt.Errorf("invalid signature")
	}

	return nil
}
