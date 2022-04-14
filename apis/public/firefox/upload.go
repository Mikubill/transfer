package firefox

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
)

func (b ffsend) Upload(files []string) {
	for _, v := range files {
		b.initUpload([]string{v})
	}
}

func (b ffsend) initUpload(files []string) {

	for _, v := range files {
		if utils.IsExist(v) {
			err := filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				err = b.upload(path)
				if err != nil {
					fmt.Printf("upload returns error: %v, onfile: %s\n", err, path)
				}
				return nil
			})
			if err != nil {
				fmt.Printf("filepath.walk(upload) returns error: %v, onfile: %s\n", err, v)
			}
		} else {
			fmt.Printf("%s not found\n", v)
		}
	}
}

func (b ffsend) upload(v string) error {
	fmt.Printf("Local: %s\n", v)
	if apis.DebugMode {
		log.Println("generating keys...")
	}

	info, err := os.Stat(v)
	if err != nil {
		return err
	}

	fmt.Printf("encrypting data from %s...", info.Name())
	key := getKeySuite()
	getAuthHeader(key)
	utils.URLSafeEncodeByte(key.authKey)
	_, err = getEncryptMetadata(info, key)
	if err != nil {
		return err
	}
	return nil
}
