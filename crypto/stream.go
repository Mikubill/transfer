package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"math"
	"sync"
)

func CalcEncryptSize(size int64) int64 {
	if size >= 1048576 {
		blocks := int64(math.Ceil(float64(size) / float64(16)))
		pads := int64(math.Floor(float64(size) / float64(1048576)))
		return blocks*16 + pads*16
	} else {
		blocks := int64(math.Ceil(float64(size) / float64(16)))
		return blocks * 16
	}
}

func StreamEncrypt(reader io.Reader, writer io.Writer, Key string, blockSize int64, sig *sync.WaitGroup) {
	iv := bytes.Repeat([]byte{'7'}, 16)
	block, err := aes.NewCipher([]byte(Key))
	if err != nil {
		panic(err)
	}
	blockMode := cipher.NewCBCEncrypter(block, iv)
	for {
		buf := make([]byte, blockSize)
		nr, err := reader.Read(buf)
		if nr <= 0 {
			break
		}
		if err != nil && err != io.EOF {
			panic(err)
		}
		if nr > 0 {
			padded := Padding(buf[:nr], block.BlockSize())
			data := make([]byte, len(padded))
			blockMode.CryptBlocks(data, padded)
			n, err := writer.Write(data)
			if err != nil || n != len(data) {
				panic(err)
			}
		}
	}
	sig.Done()
}

func StreamDecrypt(reader io.Reader, writer io.Writer, Key string, blockSize int64, sig *sync.WaitGroup) {
	iv := bytes.Repeat([]byte{'7'}, 16)
	block, err := aes.NewCipher([]byte(Key))
	if err != nil {
		panic(err)
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	for {
		buf := make([]byte, blockSize+16)
		nr, err := io.ReadFull(reader, buf)
		if nr <= 0 {
			break
		}
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			fmt.Println(err)
			break
		}
		if nr > 0 {
			data := make([]byte, nr)
			blockMode.CryptBlocks(data, buf[:nr])
			_, err := writer.Write(unPadding(data))
			if err != nil {
				panic(err)
			}
		}
	}
	sig.Done()
}
