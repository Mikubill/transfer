package crypto

import (
	"bytes"
)

func Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	//log.Println(len(src), Padding, len(append(src, paddingText...)))
	return append(src, paddingText...)
}

func unPadding(src []byte) []byte {
	n := len(src)
	//log.Println(n, int(src[n-1]), len(src[:n-int(src[n-1])]))
	return src[:n-int(src[n-1])]
}
