package crypto

import (
	"bytes"
)

func Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, paddingText...)
}

func unPadding(src []byte) []byte {
	n := len(src)
	return src[:n-int(src[n-1])]
}
