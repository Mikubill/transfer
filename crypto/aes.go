package crypto

import (
	"crypto/aes"
	"crypto/cipher"
)

func encryptAESCBC(plainText []byte, key, iv []byte) []byte {
	block, _ := aes.NewCipher(key)
	plainText = Padding(plainText, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cipherText, plainText)
	return cipherText
}

func decryptAESCBC(cipherText, key, iv []byte) []byte {
	block, _ := aes.NewCipher(key)
	blockMode := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	blockMode.CryptBlocks(plainText, cipherText)
	plainText = unPadding(plainText)
	return plainText
}

func encryptAESECB(plainText, key []byte) []byte {
	block, _ := aes.NewCipher(key)
	plainText = Padding(plainText, block.BlockSize())
	cipherText := make([]byte, len(plainText))
	size := block.BlockSize()
	for bs, be := 0, size; bs < len(cipherText); bs, be = bs+size, be+size {
		block.Encrypt(cipherText[bs:be], plainText[bs:be])
	}
	return cipherText
}

func decryptAESECB(cipherText, key []byte) []byte {
	block, _ := aes.NewCipher(key)
	plainText := make([]byte, len(cipherText))
	size := block.BlockSize()
	for bs, be := 0, size; bs < len(cipherText); bs, be = bs+size, be+size {
		block.Decrypt(plainText[bs:be], cipherText[bs:be])
	}
	return unPadding(plainText)
}
