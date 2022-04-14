package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
)

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsExist returns true if given path is exist.
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

// IsDir returns true if given path is a folder.
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsDir returns true if given path isn't folder.
func IsFile(path string) bool {
	return !IsDir(path)
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

// GenRandBytes generates a random bytes slice in given length.
func GenRandBytes(byteLength int) []byte {
	b := make([]byte, byteLength)
	_, err := rand.Read(b)
	if err != nil {
		return nil
	}
	return b
}

func GetType(v any) string {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

// GenRandString generates a random string in given length.
func GenRandString(byteLength int) (uuid string) {
	return hex.EncodeToString(GenRandBytes(byteLength))
}

// GenRandUUID generates a random uuid.
func GenRandUUID() string {
	s := GenRandString(16)
	return strings.Join([]string{s[:8], s[8:12], s[12:16], s[16:20], s[20:]}, "-")
}

// DotTicker starts a infinity dot animation and returns a chan to stop.
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

// Spacer fixed space between command and description.
func Spacer(v string) string {
	r := strings.Split(v, ":")[0]
	block := strings.Repeat(" ", 24-len(r))
	return strings.Replace(v, ":", ":"+block, 1)
}
