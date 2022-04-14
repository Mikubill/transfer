package firefox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Mikubill/transfer/utils"
	"golang.org/x/crypto/hkdf"
)

func genDeriveBytes(master []byte, info []byte) []byte {
	key := make([]byte, 16)
	b := hkdf.New(sha256.New, master, []byte(""), info)
	_, err := b.Read(key)
	if err != nil {
		return nil
	}
	return key
}

func getKeySuite() *keySuite {
	// no password only

	secretKey := utils.GenRandBytes(16)

	encryptKey := genDeriveBytes(secretKey, []byte("encrypt")) // todo: ?
	encryptIV := utils.GenRandBytes(12)                        // todo:?

	authKey := genDeriveBytes(secretKey, []byte("authentication")) // sign
	metaKey := genDeriveBytes(secretKey, []byte("metadata"))       // decrypt, encrypt
	// todo: change it to random one
	metaIV := []byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}

	return &keySuite{
		nonce:      "yRCdyQ1EMSA3mo4rqSkuNQ==",
		secretKey:  secretKey,
		encryptKey: encryptKey,
		encryptIV:  encryptIV,
		authKey:    authKey,
		metaKey:    metaKey,
		metaIV:     metaIV,
	}
}

func getAuthHeader(key *keySuite) string {
	h := hmac.New(sha256.New, key.authKey)
	h.Write([]byte(key.nonce))
	mac := string(utils.URLSafeEncodeByte(h.Sum(nil)))
	return fmt.Sprintf("send-v1 %s", mac)
}

func getEncryptMetadata(info os.FileInfo, key *keySuite) ([]byte, error) {
	metadata, err := json.Marshal(map[string][]byte{
		"iv":   utils.URLSafeEncodeByte(key.encryptIV),
		"name": []byte(info.Name()),
		"size": []byte(strconv.FormatInt(info.Size(), 10)),
		"type": []byte("application/octet-stream"),
	})
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key.metaKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	//tag := make([]byte, 12)
	cipherText := gcm.Seal(nil, key.metaIV, metadata, nil)
	return cipherText, nil
}
