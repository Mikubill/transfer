package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
	"transfer/apis"
	"transfer/utils"

	"github.com/google/uuid"
	"github.com/kjk/notionapi"
)

// Command Types
const (
	CommandSet        = "set"
	CommandUpdate     = "update"
	CommandListAfter  = "listAfter"
	CommandListRemove = "listRemove"

	signedURLPrefix = "https://www.notion.so/signed"
	s3URLPrefix     = "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/"
	notionHost      = "https://www.notion.so"
	userAgent       = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3483.0 Safari/537.36"
	acceptLang      = "en-US,en;q=0.9"
)

func PrintStruct(emp interface{}) {
	empJSON, err := json.MarshalIndent(emp, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("MarshalIndent funnction output\n %s\n", string(empJSON))
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func SubmitTransaction(ops []*Operation, spaceID string, client *Client) error {

	reqData := &submitTransactionRequest{
		RequestID: uuid.New().String(),
		Transaction: []Transaction{{
			ID:         uuid.New().String(),
			SpaceID:    spaceID,
			Operations: ops,
		}},
	}
	if apis.DebugMode {
		PrintStruct(reqData)
	}
	var rsp map[string]interface{}
	data, err := doNotionAPI(client, "/api/v3/saveTransactions", reqData, &rsp)
	if apis.DebugMode {
		PrintStruct(data)
	}
	return err
}

func closeNoError(c io.Closer) {
	_ = c.Close()
}

func doNotionAPI(c *Client, apiURL string, requestData interface{}, result interface{}) (map[string]interface{}, error) {
	var js []byte
	var err error
	if requestData != nil {
		js, err = json.Marshal(requestData)
		if err != nil {
			return nil, err
		}
	}
	uri := notionHost + apiURL
	body := bytes.NewBuffer(js)
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", acceptLang)
	if c.AuthToken != "" {
		req.Header.Set("cookie", fmt.Sprintf("token_v2=%v", c.AuthToken))
	}
	var rsp *http.Response

	rsp, err = http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer closeNoError(rsp.Body)

	if rsp.StatusCode != 200 {
		_, _ = ioutil.ReadAll(rsp.Body)
		return nil, fmt.Errorf("http.Post('%s') returned non-200 status code of %d", uri, rsp.StatusCode)
	}
	d, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(d, result)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(d, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// buildOp creates an Operation for this block
func buildOp(blockID, command string, path []string, args interface{}) *Operation {
	return &Operation{
		Point: Pointer{
			ID:    blockID,
			Table: "block",
		},
		Path:    path,
		Command: command,
		Args:    args,
	}
}

func (b *notion) DoUpload(name string, size int64, file io.Reader) error {

	if b.pageID == "" || b.token == "" {
		return fmt.Errorf("invalid pageid or token")
	}
	client := &Client{notionapi.Client{AuthToken: b.token}}
	page, err := client.DownloadPage(b.pageID)
	if err != nil {
		log.Fatalf("DownloadPage() failed with %s\n", err)
	}

	root := page.BlockByID(page.ID)
	if apis.DebugMode {
		PrintStruct(root)
	}

	fileID, fileURL, err := client.UploadFile(file, name, size)
	if err != nil {
		log.Fatalf("DownloadPage() failed with %s\n", err)
	}

	var lastBlockID string
	if len(root.Content) > 0 {
		lastBlockID = root.Content[len(root.Content)-1].ID
	}

	userID := root.LastEditedByID
	spaceID := root.ParentID
	if b.spaceID != "" {
		spaceID = b.spaceID
	}
	newBlockID := uuid.New().String()
	fmt.Printf("syncing blocks..")
	end := utils.DotTicker()

	ops := []*Operation{
		buildOp(newBlockID, CommandSet, []string{}, map[string]interface{}{
			"type":    "file",
			"id":      newBlockID,
			"version": 1,
		}),
		buildOp(newBlockID, CommandUpdate, []string{}, map[string]interface{}{
			"parent_id":    root.ID,
			"parent_table": "block",
			"alive":        true,
		}),
		buildOp(root.ID, CommandListAfter, []string{"content"}, map[string]string{
			"id":    newBlockID,
			"after": lastBlockID,
		}),
		buildOp(newBlockID, CommandSet, []string{"created_by_id"}, userID),
		buildOp(newBlockID, CommandSet, []string{"created_by_table"}, "notion_user"),
		buildOp(newBlockID, CommandSet, []string{"created_time"}, time.Now().UnixNano()),
		buildOp(newBlockID, CommandSet, []string{"last_edited_time"}, time.Now().UnixNano()),
		buildOp(newBlockID, CommandSet, []string{"last_edited_by_id"}, userID),
		buildOp(newBlockID, CommandSet, []string{"last_edited_by_table"}, "notion_user"),
		buildOp(newBlockID, CommandUpdate, []string{"properties"}, map[string]interface{}{
			"source": [][]string{{fileURL}},
			"size":   [][]string{{ByteCountIEC(size)}},
			"title":  [][]string{{name}},
		}),
		buildOp(newBlockID, CommandListAfter, []string{"file_ids"}, map[string]string{
			"id": fileID,
		}),
	}

	SubmitTransaction(ops, spaceID, client)
	*end <- struct{}{}
	fmt.Printf("%s\n", newBlockID)
	b.resp = fmt.Sprintf("%s/%s?table=block&id=%s&name=%s&userId=%s&cache=v2", signedURLPrefix, url.QueryEscape(fileURL), newBlockID, name, userID)
	return nil
}

// getUploadFileURL executes a raw API call: POST /api/v3/getUploadFileUrl
func (c *Client) getUploadFileURL(name, contentType string) (*GetUploadFileUrlResponse, error) {
	const apiURL = "/api/v3/getUploadFileUrl"

	req := &getUploadFileUrlRequest{
		Bucket:      "secure",
		ContentType: contentType,
		Name:        name,
	}

	var rsp GetUploadFileUrlResponse
	var err error
	rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp)
	if err != nil {
		return nil, err
	}

	rsp.Parse()

	return &rsp, nil
}

func (r *GetUploadFileUrlResponse) Parse() {
	r.FileID = strings.Split(r.URL[len(s3URLPrefix):], "/")[0]
}

// UploadFile Uploads a file to notion's asset hosting(aws s3)
func (c *Client) UploadFile(file io.Reader, name string, size int64) (fileID, fileURL string, err error) {
	ext := path.Ext(name)
	mt := mime.TypeByExtension(ext)
	if mt == "" {
		mt = "application/octet-stream"
	}
	// 1. getUploadFileURL
	uploadFileURLResp, err := c.getUploadFileURL(name, mt)
	if err != nil {
		err = fmt.Errorf("get upload file URL error: %s", err)
		return
	}

	// 2. Upload file to amazon - PUT
	httpClient := http.DefaultClient

	req, err := http.NewRequest(http.MethodPut, uploadFileURLResp.SignedPutURL, file)
	if err != nil {
		return
	}
	req.ContentLength = size
	req.TransferEncoding = []string{"identity"} // disable chunked (unsupported by aws)
	req.Header.Set("Content-Type", mt)
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var contents []byte
		contents, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			contents = []byte(fmt.Sprintf("Error from ReadAll: %s", err))
		}

		err = fmt.Errorf("http PUT '%s' failed with status %s: %s", req.URL, resp.Status, string(contents))
		return
	}

	return uploadFileURLResp.FileID, uploadFileURLResp.URL, nil
}

func (b notion) PostUpload(string, int64) (string, error) {
	fmt.Printf("Download Link: %s\n", b.resp)
	return b.resp, nil
}
