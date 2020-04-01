package crypto

import "bytes"

func padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, paddingText...)
}

func unPadding(src []byte) []byte {
	n := len(src)
	unPadding := int(src[n-1])
	return src[:n-unPadding]
}
