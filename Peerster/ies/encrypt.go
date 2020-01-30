//encrypt contains method to encrypt and decrypt packets. @author Johan Lanzrein

package ies

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

//KeyPair a pair of public and private key
type KeyPair struct {
	PublicKey
	PrivateKey
}

type PrivateKey []byte
type PublicKey []byte

const IVLEN = 16

//Encrypt encrypts a msg with key. returns the encrypted packet.
func Encrypt(key []byte, msg []byte) []byte {
	trZeros := TrailingZeros(msg)
	blockcipher, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	iv := make([]byte, IVLEN)
	cnt, err := rand.Read(iv)
	if cnt != IVLEN || err != nil {
		panic(err)
	}
	cbc := cipher.NewCBCEncrypter(blockcipher, iv)
	padding := aes.BlockSize - (len(msg) % aes.BlockSize)

	ciphertext := make([]byte, len(msg)+padding)
	pt := make([]byte, len(msg)+padding)
	copy(pt, msg)
	cbc.CryptBlocks(ciphertext, pt)
	var buf bytes.Buffer
	buf.Write(iv)
	buf.WriteByte(trZeros)
	buf.Write(ciphertext)
	return buf.Bytes()

}

//Computes trailing zeros of a message. max 255
func TrailingZeros(msg []byte) byte {
	cnt := byte(0)
	for i := len(msg) - 1; i >= 0; i-- {
		if msg[i] != 0 {
			return cnt
		}
		cnt++
	}
	return 255
}

//Decrypt decrypts a msg with the key returns the decrypted msg
func Decrypt(key []byte, msg []byte) []byte {
	iv := msg[0:IVLEN]
	trZeros := msg[IVLEN]
	ct := msg[IVLEN+1:]
	blockcipher, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	pt := make([]byte, len(ct))
	cbc := cipher.NewCBCDecrypter(blockcipher, iv)
	cbc.CryptBlocks(pt, ct)
	pt = bytes.TrimRight(pt, "\x00")
	zeros := make([]byte, trZeros)
	return append(pt, zeros...)
}
