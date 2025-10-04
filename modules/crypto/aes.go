package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type AesGcm struct {
	aead cipher.AEAD
}

type SharedSecret [32]byte

func (s SharedSecret) String() string {
	return base64.RawURLEncoding.EncodeToString(s[:])
}

func SecretFromString(s string) (SharedSecret, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil || len(b) != 32 {
		return SharedSecret{}, fmt.Errorf(`bad aes key`)
	}
	var k SharedSecret
	copy(k[:], b)
	return k, nil
}

func NewSecret() SharedSecret {
	var s SharedSecret
	rand.Read(s[:])
	return s
}

func NewAesGcm(sharedSecret SharedSecret) (*AesGcm, error) {
	block, err := aes.NewCipher(sharedSecret[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AesGcm{aead: gcm}, nil
}

// data will be overwritten.
// 如果 nonce 不为空，一定为 12 字节随机数据。
func (b *AesGcm) Encrypt(data []byte, nonce []byte) ([]byte, error) {
	if len(nonce) == 0 {
		nonce = make([]byte, b.aead.NonceSize())
		rand.Read(nonce)
	}
	if len(nonce) != b.aead.NonceSize() {
		return nil, fmt.Errorf(`bad aes nonce`)
	}
	sealed := b.aead.Seal(data[:0], nonce, data, nil)
	return append(sealed, nonce...), nil
}

// data will be overwritten.
func (b *AesGcm) Decrypt(data []byte) ([]byte, error) {
	if len(data) < b.aead.NonceSize() {
		return nil, fmt.Errorf(`invalid data to decrypt`)
	}
	nonce := make([]byte, b.aead.NonceSize())
	pos := len(data) - b.aead.NonceSize()
	if n := copy(nonce, data[pos:]); n != b.aead.NonceSize() {
		panic(`internal error: nonce size mismatch`)
	}
	return b.aead.Open(data[:0], nonce, data[:pos], nil)
}
