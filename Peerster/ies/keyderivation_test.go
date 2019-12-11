package ies

import (
	"crypto/subtle"
	"go.dedis.ch/onet/log"
	"testing"
	"time"
)

func TestKeyPair_KeyDerivation(t *testing.T) {
	now := time.Now()

	kp1, err := GenerateKeyPair()
	if err != nil{
		t.Fatal(err)
	}
	kp2, err := GenerateKeyPair()
	if err != nil{
		t.Fatal(err)
	}
	elapsed := time.Since(now)
	log.Lvl1("Two key pair generated in :" , elapsed.Microseconds(),"us")
	now = time.Now()
	shared1 := kp1.KeyDerivation(&kp2.PublicKey)
	shared2 := kp2.KeyDerivation(&kp1.PublicKey)
	elapsed = time.Since(now)
	log.Lvl1("Two shared key derived in : ", elapsed.Microseconds(),"us")

	log.Lvl1("Comparing shared keys...")
	now = time.Now()
	if subtle.ConstantTimeCompare(shared1,shared2) != 1 {
		t.Fatal("Error : shared keys do not match")
	}

	elapsed = time.Since(now)
	log.Lvl1("Shared keys are equal.\nCompared keys in :", elapsed)


}


func TestEncryptDecrypt(t *testing.T){
	now := time.Now()

	kp1, err := GenerateKeyPair()
	if err != nil{
		t.Fatal(err)
	}
	kp2, err := GenerateKeyPair()
	if err != nil{
		t.Fatal(err)
	}
	elapsed := time.Since(now)
	log.Lvl1("Two key pair generated in :" , elapsed.Microseconds(),"us")
	now = time.Now()
	shared1 := kp1.KeyDerivation(&kp2.PublicKey)
	shared2 := kp2.KeyDerivation(&kp1.PublicKey)
	elapsed = time.Since(now)
	log.Lvl1("Two shared key derived in : ", elapsed.Microseconds(),"us")

	log.Lvl1("Comparing shared keys...")
	now = time.Now()
	if subtle.ConstantTimeCompare(shared1,shared2) != 1 {
		t.Fatal("Error : shared keys do not match")
	}

	elapsed = time.Since(now)
	log.Lvl1("Shared keys are equal.\nCompared keys in :", elapsed)

	msg := "Hello peerster ! "
	cipher := Encrypt(shared1, []byte(msg))

	log.Lvlf1("Cipher text : %x", cipher )
	pt := Decrypt(shared2, cipher)
	log.Lvl1("Resulting cleartext : " , string(pt))
	log.Lvlf1("original : %x, pt : %x ", []byte(msg), pt)
	//todo here it says that we do not have matching ciphers. but they are...
	if subtle.ConstantTimeCompare(pt, []byte(msg)) != 1{
		t.Fatal("Error resulting plaintext does not match ! " )
	}
}
