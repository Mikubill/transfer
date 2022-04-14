package crypto

import (
	"bytes"
	"testing"

	"github.com/Mikubill/transfer/utils"
)

func TestDES(t *testing.T) {
	raw := utils.GenRandBytes(16)
	key := utils.GenRandBytes(8)
	src, err := encryptDES(raw, key)
	if err != nil {
		t.Fatal("des: failed", err)
	}
	dec, err := decryptDES(src, key)
	if err != nil {
		t.Fatal("des: failed", err)
	}
	if bytes.Equal(dec, raw) {
		t.Log("des: success")
	} else {
		t.Fatal("des: failed", err)
	}
}
