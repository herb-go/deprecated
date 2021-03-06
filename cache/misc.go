package cache

import (
	"bytes"
	"crypto/rand"
)

//TokenMask The []bytes of alphabet and number to generate token.
var TokenMask = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_.")

//RandomBytes Generate a give length random []byte.
//Return the random [] byte and any error raised.
func RandomBytes(length int) ([]byte, error) {
	token := make([]byte, length)
	_, err := rand.Read(token)
	return token, err
}

//NewRandomBytes Generate a give length random []byte which different from origin bytes.
//Return the random [] byte and any error raised.
func NewRandomBytes(length int, origin []byte) ([]byte, error) {
	for {
		token, err := RandomBytes(length)
		if err != nil {
			return token, err
		}
		if bytes.Compare(token, origin) != 0 {
			return token, nil
		}
	}
}

//RandMaskedBytes Generate a give length random []byte.
//All bytes in the random []byte is select from given mask.
//Return the random [] byte and any error raised.
func RandMaskedBytes(mask []byte, length int) ([]byte, error) {
	token := make([]byte, length)
	masked := make([]byte, length)
	_, err := rand.Read(token)
	if err != nil {
		return masked, err
	}
	l := len(mask)
	for k, v := range token {
		index := int(v) % l
		masked[k] = mask[index]
	}
	return masked, nil
}

//NewRandMaskedBytes Generate a give length random []byte which different from origin bytes.
//All bytes in the random []byte is select from given mask.
//Return the random [] byte and any error raised.
func NewRandMaskedBytes(mask []byte, length int, origin []byte) ([]byte, error) {
	for {
		token, err := RandMaskedBytes(mask, length)
		if err != nil {
			return token, err
		}
		if bytes.Compare(token, origin) != 0 {
			return token, nil
		}
	}
}
