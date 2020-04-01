package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"os"
	"strings"
	"time"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		//log.Println(err)
		return false
	}
	return true
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func IsFile(path string) bool {
	return !IsDir(path)
}

func HashBlock(buf []byte) int {
	return int(crc32.ChecksumIEEE(buf))
}

func URLSafeEncodeByte(enc []byte) []byte {
	r := make([]byte, base64.StdEncoding.EncodedLen(len(enc)))
	base64.StdEncoding.Encode(r, enc)
	r = bytes.ReplaceAll(r, []byte("+"), []byte("-"))
	r = bytes.ReplaceAll(r, []byte("/"), []byte("_"))
	return r
}

func URLSafeEncode(enc string) string {
	r := base64.StdEncoding.EncodeToString([]byte(enc))
	r = strings.ReplaceAll(r, "+", "-")
	r = strings.ReplaceAll(r, "/", "_")
	return r
}

func GetFileInfo(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func GenRandBytes(byteLength int) []byte {
	b := make([]byte, byteLength)
	_, err := rand.Read(b)
	if err != nil {
		return nil
	}
	return b
}

func GenRandString(byteLength int) (uuid string) {
	return hex.EncodeToString(GenRandBytes(byteLength))
}

func GenRandUUID() string {
	s := GenRandString(16)
	return strings.Join([]string{s[:8], s[8:12], s[12:16], s[16:20], s[20:]}, "-")
}

func DotTicker() *chan struct{} {
	tick := time.NewTicker(time.Second)
	end := make(chan struct{})
	go func() {
		for {
			select {
			case <-tick.C:
				fmt.Printf(".")
			case <-end:
				return
			}
		}
	}()
	return &end
}

func Spacer(v string) string {
	r := strings.Split(v, ":")[0]
	block := strings.Repeat(" ", 24-len(r))
	return strings.Replace(v, ":", ":"+block, 1)
}
