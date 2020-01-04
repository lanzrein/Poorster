package ies

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

type KeyPair struct {
	PublicKey
	PrivateKey
}

type PrivateKey []byte
type PublicKey []byte

const IVLEN = 16

func Encrypt(key []byte, msg []byte) []byte {
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
	buf.Write(ciphertext)
	return buf.Bytes()

}

func Decrypt(key []byte, msg []byte) []byte {
	iv := msg[0:IVLEN]
	ct := msg[IVLEN:]
	blockcipher, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	pt := make([]byte, len(ct))
	cbc := cipher.NewCBCDecrypter(blockcipher, iv)
	cbc.CryptBlocks(pt, ct)
	pt = bytes.TrimRight(pt, "\x00")

	return pt
}
