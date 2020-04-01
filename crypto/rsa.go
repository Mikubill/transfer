package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
)

func rsaEncrypt(originalData, publicKey []byte) ([]byte, error) {
	pubKey, _ := x509.ParsePKIXPublicKey(publicKey)
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey.(*rsa.PublicKey), originalData)
	return encryptedData, err
}

func rsaDecrypt(encryptedData, privateKey []byte) ([]byte, error) {
	prvKey, _ := x509.ParsePKCS1PrivateKey(privateKey)
	originalData, err := rsa.DecryptPKCS1v15(rand.Reader, prvKey, encryptedData)
	return originalData, err
}
