package gofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
)

const (
	getServer              = "https://api.gofile.io/getServer"
	createAccount          = "https://api.gofile.io/createAccount"
	createFolder           = "https://api.gofile.io/createFolder"
	setFolder              = "https://api.gofile.io/setOption"
	getUserAccount         = "https://api.gofile.io/getAccountDetails?token=%s"
	createFolderPostString = "parentFolderId=%s&folderName=%s&token=%s"
	setPrimPostString      = "contentId=%s&token=%s&option=public&value=true"
	shareFolder            = "https://api.gofile.io/shareFolder?folderId=%s&token=%s"
)

func (b *goFile) InitUpload(_ []string, sizes []int64) error {
	// createAccount
	fmt.Printf("fetching UserToken and FolderToken..")
	end := utils.DotTicker()

	err := b.createUser()
	if err != nil {
		return err
	}

	err = b.createFolder()
	if err != nil {
		return err
	}

	*end <- struct{}{}
	fmt.Printf("done\n")

	err = b.selectServer()
	if err != nil {
		return err
	}

	return nil
}

func smallParser(body *http.Response, result any) error {
	data, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return fmt.Errorf("read body returns error: %v", err)
	}

	if apis.DebugMode {
		log.Printf("parsing json: %s", string(data))
	}

	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("parse body returns error: %v", err)
	}

	if apis.DebugMode {
		log.Printf("parsed data: %+v", result)
	}
	return nil
}

func (b *goFile) createFolder() error {

	// first check root folderId
	if apis.DebugMode {
		log.Printf("#1 check root folderId")
	}
	getUserDetails := fmt.Sprintf(getUserAccount, b.userToken)
	body, err := http.Get(getUserDetails)
	if err != nil {
		return fmt.Errorf("request %s returns error: %v", getServer, err)
	}

	var sevData1 userDetails
	err = smallParser(body, &sevData1)
	if err != nil {
		return fmt.Errorf("request %s returns error: %v", getServer, err)
	}

	body.Body.Close()
	rootFolderID := sevData1.Data.RootFolder
	postString := fmt.Sprintf(createFolderPostString, rootFolderID, "tfolder", b.userToken)

	// then create new folder
	if apis.DebugMode {
		log.Printf("#2 create new folder")
	}
	var sevData2 folderDetails
	req, err := http.NewRequest("PUT", createFolder, strings.NewReader(postString))
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}
	if err = reqSender(req, &sevData2); err != nil {
		return err
	}
	b.folderID = sevData2.Data.ID
	b.folderName = sevData2.Data.Name
	// b.userToken = sevData2.Data.Token

	// set folder as public
	postString = fmt.Sprintf(setPrimPostString, b.folderID, b.userToken)
	req, err = http.NewRequest("PUT", setFolder, strings.NewReader(postString))
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}
	if err = reqSender(req, &sevData2); err != nil {
		return err
	}

	return nil
}

func reqSender(req *http.Request, parsed any) error {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	body, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s returns error: %v", getServer, err)
	}
	err = smallParser(body, parsed)
	if err != nil {
		return fmt.Errorf("request %s returns error: %v", getServer, err)
	}
	body.Body.Close()
	return nil
}

func (b *goFile) createUser() error {

	body, err := http.Get(createAccount)
	if err != nil {
		return fmt.Errorf("request %s returns error: %v", getServer, err)
	}
	var sevData userGen
	err = smallParser(body, &sevData)
	if err != nil {
		return fmt.Errorf("request %s returns error: %v", getServer, err)
	}
	body.Body.Close()
	b.userToken = sevData.Data.Token

	return nil
}

func (b *goFile) selectServer() error {

	fmt.Printf("Selecting server..")
	end := utils.DotTicker()
	body, err := http.Get(getServer)
	if err != nil {
		return fmt.Errorf("request %s returns error: %v", getServer, err)
	}

	data, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return fmt.Errorf("read body returns error: %v", err)
	}
	_ = body.Body.Close()

	var sevData respBody
	if err := json.Unmarshal(data, &sevData); err != nil {
		return fmt.Errorf("parse body returns error: %v", err)
	}
	*end <- struct{}{}
	fmt.Printf("%s\n", strings.TrimSpace(sevData.Data.Server))
	// srv-store0 has dns problem
	if sevData.Data.Server == "srv-store0" {
		sevData.Data.Server = "srv-store1"
	}
	b.serverLink = fmt.Sprintf("https://%s.gofile.io/uploadFile", strings.TrimSpace(sevData.Data.Server))

	return nil
}

func (b *goFile) DoUpload(name string, size int64, file io.Reader) error {

	body, err := b.newMultipartUpload(uploadConfig{
		fileSize:   size,
		fileName:   name,
		fileReader: file,
		debug:      apis.DebugMode,
	})

	// Get download link from response
	var respData uploadResp
	err = json.Unmarshal(body, &respData)
	if err != nil {
		if apis.DebugMode {
			log.Printf("parse response error: %v", err)
		}
		return err
	}

	b.downloadLink = respData.Data.DownLoadPage
	if err != nil {
		return fmt.Errorf("upload returns error: %s", err)
	}

	return nil
}

func (b *goFile) newMultipartUpload(config uploadConfig) ([]byte, error) {
	if config.debug {
		log.Printf("start upload")
	}
	client := http.Client{}

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	_ = writer.WriteField("token", b.userToken)
	_ = writer.WriteField("folderId", b.folderID)
	_, err := writer.CreateFormFile("file", config.fileName)
	if err != nil {
		return nil, err
	}

	writerLength := byteBuf.Len()
	writerBody := make([]byte, writerLength)
	_, _ = byteBuf.Read(writerBody)
	_ = writer.Close()

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

	req, err := http.NewRequest("POST", b.serverLink, partR)
	if err != nil {
		return nil, err
	}
	req.ContentLength = totalSize
	req.Header.Set("content-length", strconv.FormatInt(totalSize, 10))
	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
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

	_ = resp.Body.Close()
	if config.debug {
		log.Printf("returns: %v", string(body))
	}

	return body, nil
}

func (b *goFile) FinishUpload([]string) (string, error) {

	folderProperties := fmt.Sprintf(shareFolder, b.folderID, b.userToken)

	body, err := http.Get(folderProperties)
	if err != nil {
		return "", fmt.Errorf("request %s returns error: %v", folderProperties, err)
	}
	var sevData sharedDetails
	err = smallParser(body, &sevData)
	if sevData.Status != "ok" {
		return "", fmt.Errorf("parse %s returns error: %v", folderProperties, err)
	}
	body.Body.Close()
	if sevData.Status != "ok" {
		return "", fmt.Errorf("request %s returns non-ok status: %v", folderProperties, sevData)
	}

	link := fmt.Sprintf("https://gofile.io/?c=%s", b.folderName)
	// fmt.Printf("Download Link: %s\nUser Token: %s\n", link, b.userToken)
	fmt.Printf("Download Link: %s\nUser Token: %s\n", b.downloadLink, b.userToken)

	return link, nil

}
