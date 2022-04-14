package anonfiles

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/apis/methods"
)

const upload = "https://api.anonfiles.com/upload"

func (b *anon) DoUpload(name string, size int64, file io.Reader) error {

	body, err := methods.MultipartUpload(methods.MultiPartUploadConfig{
		FileSize:   size,
		FileName:   name,
		FileReader: file,
		Debug:      apis.DebugMode,
		Endpoint:   upload,
	})
	if err != nil {
		return fmt.Errorf("upload returns error: %s", err)
	}

	var resp uploadResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return fmt.Errorf("unmarshal returns error: %s", err)
	}

	if !resp.Status {
		return fmt.Errorf("upload returns error: %s", resp.Error.Message)
	}

	b.resp = resp.Data.File.URL.Full
	return nil
}

func (b *anon) PostUpload(string, int64) (string, error) {
	fmt.Printf("Download Link: %s\n", b.resp)
	return b.resp, nil
}
