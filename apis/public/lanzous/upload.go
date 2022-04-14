package lanzous

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"path"
	"strconv"
	"strings"
)

const upload = "https://pc.woozooo.com/fileup.php"

type lzResp struct {
	Info string      `json:"info"`
	Text []fileBlock `json:"text"`
}

type fileBlock struct {
	ID string `json:"f_id"`
}

func (b *lanzous) DoUpload(name string, size int64, file io.Reader) error {
	if b.Config.token == "" {
		return fmt.Errorf("no token")
	}

	body, err := b.newMultipartUpload(uploadConfig{
		fileSize:   size,
		fileName:   name,
		fileReader: file,
		debug:      apis.DebugMode,
	})
	if err != nil {
		return fmt.Errorf("upload returns error: %s", err)
	}
	var resp lzResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf(string(body))
	}
	if uniToC(resp.Info) != `上传成功` {
		return fmt.Errorf(uniToC(resp.Info))
	}
	b.resp = fmt.Sprintf("https://www.lanzous.com/%s\n", resp.Text[0].ID)

	return nil
}

func (b *lanzous) PostUpload(string, int64) (string, error) {
	fmt.Printf("Download Link: %s\n", b.resp)
	return b.resp, nil
}

func (b lanzous) newMultipartUpload(config uploadConfig) ([]byte, error) {
	if config.debug {
		log.Printf("start upload")
	}
	client := http.Client{}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_ = writer.WriteField("task", "1")
	_ = writer.WriteField("folder_id", "-1")
	_ = writer.WriteField("ve", "1")
	_ = writer.WriteField("id", "WU_FILE_0")
	_ = writer.WriteField("name", extChecker(config.fileName))
	_ = writer.WriteField("type", "application/octet-stream")
	_ = writer.WriteField("size", strconv.FormatInt(config.fileSize, 10))
	_, err := writer.CreateFormFile("upload_file", extChecker(config.fileName))
	if err != nil {
		return nil, err
	}

	writerLength := byteBuf.Len()
	writerBody := make([]byte, writerLength)
	_, _ = byteBuf.Read(writerBody)
	_ = writer.Close()
	//log.Println(string(writerBody))

	boundary := byteBuf.Len()
	lastBoundary := make([]byte, boundary)
	_, _ = byteBuf.Read(lastBoundary)

	totalSize := int64(writerLength) + config.fileSize + int64(boundary)
	partR, partW := io.Pipe()

	go func() {
		_, _ = partW.Write(writerBody)
		for {
			buf := make([]byte, 256)
			nr, err := io.ReadFull(config.fileReader, buf)
			if nr <= 0 {
				break
			}
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				fmt.Println(err)
				break
			}
			if nr > 0 {
				_, _ = partW.Write(buf[:nr])
			}
		}
		_, _ = partW.Write(lastBoundary)
		_ = partW.Close()
	}()
	var req *http.Request
	req, err = http.NewRequest("POST", upload, partR)

	if err != nil {
		return nil, err
	}
	req.ContentLength = totalSize
	req.Header.Set("content-length", strconv.FormatInt(totalSize, 10))
	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	req.Header.Set("cookie", b.Config.token)
	if config.debug {
		log.Printf("header: %v", req.Header)
	}

	resp, err := client.Do(req)
	if err != nil {
		if config.debug {
			log.Printf("do requests returns error: %v", err)
		}
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.debug {
			log.Printf("read response returns: %v", err)
		}
		return nil, err
	}
	//log.Println(string(body))
	_ = resp.Body.Close()
	if config.debug {
		log.Printf("returns: %v", string(body))
	}

	return body, nil
}

var validExt = []string{"doc", "docx", "zip", "rar", "apk", "ipa", "txt",
	"exe", "7z", "e", "z", "ct", "ke", "cetrainer", "db", "tar", "pdf",
	"w3x", "epub", "mobi", "azw", "azw3", "osk", "osz", "xpa", "cpk",
	"lua", "jar", "dmg", "ppt", "pptx", "xls", "xlsx", "mp3", "ipa",
	"iso", "img", "gho", "ttf", "ttc", "txf", "dwg", "bat", "imazingapp",
	"dll", "crx", "xapk", "conf", "deb", "rp", "rpm", "rplib",
	"mobileconfig", "appimage", "lolgezi", "flac"}

func extChecker(n string) string {
	_ex := path.Ext(n)
	for _, ext := range validExt {
		if strings.HasSuffix(_ex, ext) {
			return n
		}
	}
	return n + ".zip"
}

func uniToC(u string) string {
	textQuoted := strconv.QuoteToASCII(u)
	textUnquoted := textQuoted[1 : len(textQuoted)-1]
	ulnar := strings.Split(textUnquoted, "\\u")
	var context string
	for _, v := range ulnar {
		if len(v) < 1 {
			continue
		}
		temp, err := strconv.ParseInt(v, 16, 32)
		if err != nil {
			panic(err)
		}
		context += fmt.Sprintf("%c", temp)
	}
	return context
}
