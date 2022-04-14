package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"github.com/Mikubill/transfer/utils"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func Hash(v []string) {
	for _, item := range v {
		if utils.IsExist(item) && !utils.IsDir(item) {
			hash(item)
		}
	}
}

func hash(file string) {
	stat, _ := os.Stat(file)
	path, _ := filepath.Abs(file)
	fmt.Println("size: " + strconv.FormatInt(stat.Size(), 10))
	fmt.Println("path: " + path)
	fmt.Println("")
	fmt.Println("crc32: " + c32(file))
	fmt.Println("md5: " + m5(file))
	fmt.Println("sha1: " + s1(file))
	fmt.Println("sha256: " + s256(file))
}

func s1(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return err.Error()
	}

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return err.Error()
	}
	_ = f.Close()

	return fmt.Sprintf("%x", h.Sum(nil))
}

func c32(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return err.Error()
	}

	tablePolynomial := crc32.MakeTable(0xedb88320)
	h := crc32.New(tablePolynomial)
	if _, err := io.Copy(h, f); err != nil {
		return err.Error()
	}
	_ = f.Close()

	return fmt.Sprintf("%x", h.Sum(nil))
}

func m5(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return err.Error()
	}

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return err.Error()
	}
	_ = f.Close()

	return fmt.Sprintf("%x", h.Sum(nil))
}

func s256(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return err.Error()
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err.Error()
	}
	_ = f.Close()

	return fmt.Sprintf("%x", h.Sum(nil))
}
