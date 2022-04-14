package crypto

import (
	"bytes"
	"testing"

	"github.com/Mikubill/transfer/utils"
)

func TestECB(t *testing.T) {
	raw := utils.GenRandBytes(16)
	key := utils.GenRandBytes(8)
	src := encryptAESECB(raw, key)
	dec := decryptAESECB(src, key)
	if bytes.Equal(dec, raw) {
		t.Log("aes-ecb: success")
	} else {
		t.Fatal("aes-ecb: failed")
	}
}

func TestCBC(t *testing.T) {
	raw := utils.GenRandBytes(16)
	iv := utils.GenRandBytes(16)
	key := utils.GenRandBytes(8)
	src := encryptAESCBC(raw, key, iv)
	dec := decryptAESCBC(src, key, iv)
	if bytes.Equal(dec, raw) {
		t.Log("aes-cbc: success")
	} else {
		t.Fatal("aes-cbc: failed")
	}
}
