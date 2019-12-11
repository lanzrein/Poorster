package ies

import (
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)



func GenerateKeyPair()(*KeyPair, error){

	pk := make([]byte, 32)
	sk := make([]byte, 32)
	n, err := rand.Read(sk)
	if n != 32 || err != nil{
		panic(err)
	}
	pk , err = curve25519.X25519(sk,curve25519.Basepoint)
	if n != 32 || err != nil{
		panic(err)
	}
	kp := KeyPair{pk, sk }
	return &kp , nil

}

func PublicKeyFromBytes(data[]byte) PublicKey {
	pk := make([]byte, 32)
	copy(pk, data)
	return pk

}

//KeyDerivation given an public key derive a shared secret key using the secret key of the KeyPair
func (kp *KeyPair) KeyDerivation(public *PublicKey)[]byte{
	shared := combineKeys(&kp.PrivateKey, public)
	//use the shared key to create a master key that can be used to enc dec

	masterKey := hkdf.Extract(sha256.New, shared,nil)
	return masterKey
}

func combineKeys(sk *PrivateKey, pk *PublicKey)[]byte{
	//we need to do sk*pk
	var shared []byte

	shared, err := curve25519.X25519(*sk,*pk)
	if err != nil{
		panic(err)
	}

	return shared


}