package crypto

import (
	"crypto/cipher"
	"crypto/des"
	"fmt"
)

func encryptDES(src, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	src = Padding(src, bs)
	if len(src)%bs != 0 {
		return nil, fmt.Errorf("crypto/cipher: input not full blocks")
	}
	out := make([]byte, len(src))
	dst := out
	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	return out, nil
}

func decryptDES(src, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(src))
	dst := out
	bs := block.BlockSize()
	if len(src)%bs != 0 {
		return nil, fmt.Errorf("crypto/cipher: input not full blocks")
	}
	for len(src) > 0 {
		block.Decrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	out = unPadding(out)
	return out, nil
}

func EncryptDESCBC(plainText []byte, key, iv []byte) ([]byte, error) {
	block, err := des.NewCipher(key[:8])
	if err != nil {
		return nil, err
	}
	plainText = Padding(plainText, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cipherText, plainText)
	return cipherText, nil
}

func DecryptDESCBC(cipherText, key, iv []byte) ([]byte, error) {
	block, err := des.NewCipher(key[:8])
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	blockMode.CryptBlocks(plainText, cipherText)
	plainText = unPadding(plainText)
	return plainText, nil
}
