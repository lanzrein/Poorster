package ies

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/subtle"
	"go.dedis.ch/onet/log"
	"testing"
	"time"
)

func TestKeyPair_KeyDerivation(t *testing.T) {
	now := time.Now()

	kp1, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	kp2, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(now)
	log.Lvl1("Two key pair generated in :", elapsed.Microseconds(), "us")
	now = time.Now()
	shared1 := kp1.KeyDerivation(&kp2.PublicKey)
	shared2 := kp2.KeyDerivation(&kp1.PublicKey)
	elapsed = time.Since(now)
	log.Lvl1("Two shared key derived in : ", elapsed.Microseconds(), "us")

	log.Lvl1("Comparing shared keys...")
	now = time.Now()
	if subtle.ConstantTimeCompare(shared1, shared2) != 1 {
		t.Fatal("Error : shared keys do not match")
	}

	elapsed = time.Since(now)
	log.Lvl1("Shared keys are equal.\nCompared keys in :", elapsed)

}

func TestEncryptDecrypt(t *testing.T) {
	now := time.Now()

	kp1, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	kp2, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(now)
	log.Lvl1("Two key pair generated in :", elapsed.Nanoseconds(), "us")
	now = time.Now()
	shared1 := kp1.KeyDerivation(&kp2.PublicKey)
	shared2 := kp2.KeyDerivation(&kp1.PublicKey)
	elapsed = time.Since(now)
	log.Lvl1("Two shared key derived in : ", elapsed.Nanoseconds(), "us")

	log.Lvl1("Comparing shared keys...")
	now = time.Now()
	if subtle.ConstantTimeCompare(shared1, shared2) != 1 {
		t.Fatal("Error : shared keys do not match")
	}

	elapsed = time.Since(now)
	log.Lvl1("Shared keys are equal. \nCompared keys in :", elapsed.Nanoseconds())
	for i := 0; i < 10; i++ {
		data := make([]byte, aes.BlockSize*(i+1)*100000)
		rand.Read(data)
		now = time.Now()
		cipher := Encrypt(shared1, data)
		encTime := time.Since(now)
		now = time.Now()
		log.Lvlf3("Cipher text : %x", cipher)
		pt := Decrypt(shared2, cipher)
		decTime := time.Since(now)
		log.Lvl1("For data len : ", len(data), "Enc time : ", encTime.String(), ";Dec time : ", decTime.String())
		log.Lvl3("Resulting cleartext : ", pt)
		log.Lvlf3("original : %x, pt : %x ", data, pt)
		if subtle.ConstantTimeCompare(pt, data) != 1 {
			t.Fatal("Error resulting plaintext does not match ! ")
		}
	}

}
