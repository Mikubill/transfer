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
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Mikubill/transfer/apis"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

const (
	s3URLPrefix = "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/"
	userAgent   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3483.0 Safari/537.36"
	acceptLang  = "en-US,en;q=0.9"

	CommandSet          = "set"
	CommandUpdate       = "update"
	CommandListAfter    = "listAfter"
	CommandListRemove   = "listRemove"
	TableSpace          = "space"
	TableActivity       = "activity"
	TableBlock          = "block"
	TableUser           = "notion_user"
	TableCollection     = "collection"
	TableCollectionView = "collection_view"
	TableComment        = "comment"
	TableDiscussion     = "discussion"

	dashIDLen   = len("2131b10c-ebf6-4938-a127-7089ff02dbe4")
	noDashIDLen = len("2131b10cebf64938a1277089ff02dbe4")
	webAPI      = "https://www.notion.so/api/v3"
	layout      = "2006-01-02T15:04:05.999Z"
)

func NewWebClient(token string) *webClient {
	return &webClient{
		token:   token,
		limiter: rate.NewLimiter(rate.Every(time.Second*5), 1),
	}
}

// ToDashID convert id in format bb760e2dd6794b64b2a903005b21870a
// to bb760e2d-d679-4b64-b2a9-03005b21870a
// If id is not in that format, we leave it untouched.
func ToDashID(id string) string {
	s := strings.Replace(id, "-", "", -1)
	if len(s) != noDashIDLen {
		return id
	}
	res := id[:8] + "-" + id[8:12] + "-" + id[12:16] + "-" + id[16:20] + "-" + id[20:]
	return res
}

func (c *webClient) GetPage(pageid string) (res *PageDataResponse, err error) {
	pageID := ToDashID(pageid)

	res, err = c.GetPageData(pageID)
	if err != nil {
		return
	}
	if res.Spaceid == "" || res.Owneruserid == "" {
		return nil, fmt.Errorf("metadata not found, page \"%s\"", pageID)
	}
	res.Pageid = pageID

	var cur *cursor
	var rsp *LoadPageChunkResponse
	var last *Block
	chunkID := 0
	for {
		rsp, err = c.LoadPageChunk(pageID, chunkID, cur)
		if err != nil {
			return nil, err
		}
		recordLoc := rsp.RecordMap
		if recordLoc == nil {
			chunkID++
			cur = &rsp.Cursor
			continue
		}

		for _, v := range recordLoc.Blocks {
			b := v.Block
			if b.Alive {
				last = b
			}
		}
		break
	}
	res.Cursor = last
	return
}

func (c *webClient) GetFullPage(pageid string) (last []*Block) {
	pageID := ToDashID(pageid)

	var cur *cursor
	chunkNo := 0
	for {
		rsp, err := c.LoadPageChunk(pageID, chunkNo, cur)
		if err != nil {
			// log.Warn(err.Error())
			break
		}
		chunkNo++
		recordLoc := rsp.RecordMap
		for _, v := range recordLoc.Blocks {
			b := v.Block
			if b.Alive && b.Type == "file" {
				last = append(last, b)
			}
		}
		cur = &rsp.Cursor
		if len(cur.Stack) == 0 {
			break
		}
	}
	return
}

func (c *webClient) insertFile(filename, fileid, fileurl string,
	root *PageDataResponse, filesize int64, mod time.Time) (string, error) {
	// PrintStruct(root)

	userID := root.Owneruserid
	spaceID := root.Spaceid
	timeStamp := time.Now().Unix() * 1000

	newBlockID := uuid.New().String()

	ops := []*Operation{
		buildOp(newBlockID, CommandSet, []string{}, map[string]any{
			"type":    "file",
			"id":      newBlockID,
			"version": 1,
		}),
		buildOp(newBlockID, CommandUpdate, []string{}, map[string]any{
			"parent_id":    root.Pageid,
			"parent_table": "block",
			"alive":        true,
		}),
		buildOp(root.Pageid, CommandListAfter, []string{"content"}, map[string]string{
			"id":    newBlockID,
			"after": root.Cursor.ID,
		}),
		buildOp(newBlockID, CommandSet, []string{"created_by_id"}, userID),
		buildOp(newBlockID, CommandSet, []string{"created_by_table"}, "notion_user"),
		buildOp(newBlockID, CommandSet, []string{"created_time"}, timeStamp),
		buildOp(newBlockID, CommandSet, []string{"last_edited_time"}, timeStamp),
		buildOp(newBlockID, CommandSet, []string{"last_edited_by_id"}, userID),
		buildOp(newBlockID, CommandSet, []string{"last_edited_by_table"}, "notion_user"),
		buildOp(newBlockID, CommandUpdate, []string{"properties"}, map[string]any{
			"source":          [][]string{{fileurl}},
			"size":            [][]string{{ByteCountIEC(filesize)}},
			"title":           [][]string{{filename}},
			"actual_size":     [][]string{{strconv.FormatInt(filesize, 10)}},
			"actual_modified": [][]string{{strconv.FormatInt(mod.UnixNano(), 10)}},
		}),
		buildOp(newBlockID, CommandListAfter, []string{"file_ids"}, map[string]string{
			"id": fileid,
		}),
	}
	err := c.pushTransaction(ops, newBlockID, spaceID, filename)
	if err != nil {
		return "", err
	}
	return newBlockID, nil
	// lastBlockID = newBlockID
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

// buildOp creates an Operation for this block
func buildOp(blockID, command string, path []string, args any) *Operation {
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

func (c *webClient) pushTransaction(ops []*Operation, blockID, spaceID, filename string) error {
	c.limiter.Allow()
	err := c.SubmitTransaction(ops, spaceID)
	if err != nil {
		// log.Println(err)
		return err
	}
	// log.Println("Uploaded.", filename)
	return nil
}

func (c *webClient) SubmitTransaction(ops []*Operation, spaceID string) error {

	reqData := &submitTransactionRequest{
		RequestID: uuid.New().String(),
		Transaction: []Transaction{{
			ID:         uuid.New().String(),
			SpaceID:    spaceID,
			Operations: ops,
		}},
	}
	// PrintStruct(reqData)

	js, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "https://www.notion.so/api/v3/saveTransactions", bytes.NewBuffer(js))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", acceptLang)
	req.Header.Set("cookie", fmt.Sprintf("token_v2=%v", c.token))
	var rsp *http.Response

	// http.DefaultClient.Timeout = time.Second * 30
	rsp, err = http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	d, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	_ = rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return fmt.Errorf("http.Post returned non-200 status code of %d, returns: %s", rsp.StatusCode, d)
	}

	if !bytes.Equal(d, []byte("{}")) {
		return fmt.Errorf("unknown error: %s", d)
	}
	return nil
}

// UploadFile Uploads a file to notion's asset hosting(aws s3)
func (c *webClient) UploadFile(file io.Reader, name string, size int64) (fileID, fileURL string, err error) {
	mt := mime.TypeByExtension(path.Ext(name))
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

// getUploadFileURL executes a raw API call: POST /api/v3/getUploadFileUrl
func (c *webClient) getUploadFileURL(name, contentType string) (*GetUploadFileUrlResponse, error) {
	req := &getUploadFileUrlRequest{
		Bucket:      "secure",
		ContentType: contentType,
		Name:        name,
	}

	var rsp GetUploadFileUrlResponse
	var err error
	rsp.RawJSON, err = c.doNotionAPI("/getUploadFileUrl", req, &rsp)
	if err != nil {
		return nil, err
	}

	rsp.Parse()

	return &rsp, nil
}

func (r *GetUploadFileUrlResponse) Parse() {
	r.FileID = strings.Split(r.URL[len(s3URLPrefix):], "/")[0]
}

func (c *webClient) doNotionAPI(apiURL string, requestData any, result any) (map[string]any, error) {
	var js []byte
	var err error

	if requestData != nil {
		js, err = json.Marshal(requestData)
		if err != nil {
			return nil, err
		}
	}
	uri := webAPI + apiURL
	if apis.DebugMode {
		log.Println(uri)
		log.Printf("%s", js)
	}
	body := bytes.NewBuffer(js)
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", acceptLang)
	req.Header.Set("cookie", fmt.Sprintf("token_v2=%v", c.token))
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
	if apis.DebugMode {
		log.Printf("%s", d)
	}
	// log.Prin("%s", d)
	err = json.Unmarshal(d, result)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(d, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func closeNoError(c io.Closer) {
	_ = c.Close()
}

// GetRecordValues executes a raw API call /api/v3/getRecordValues
func (c *webClient) GetPageData(pageid string) (*PageDataResponse, error) {
	req := &PageDataRequest{
		Type:       "block-space",
		Blockid:    pageid,
		Name:       "page",
		Saveparent: false,
		Showmoveto: false,
	}

	var rsp PageDataResponse
	var err error
	if _, err = c.doNotionAPI("/getPublicPageData", req, &rsp); err != nil {
		return nil, err
	}

	return &rsp, nil
}

// table is not always present in Record returned by the server
// so must be provided based on what was asked
func parseRecord(table string, r *Record) error {
	// it's ok if some records don't return a value
	if len(r.Value) == 0 {
		return nil
	}
	if r.Table == "" {
		r.Table = table
	} else {
		// TODO: probably never happens
		return fmt.Errorf("%+v != %+v", r.Table, table)
	}

	// set Block/Space etc. based on TableView type
	var pRawJSON *map[string]any
	var obj any
	switch table {
	case TableBlock:
		r.Block = &Block{}
		obj = r.Block
		pRawJSON = &r.Block.RawJSON
	}
	if obj == nil {
		return fmt.Errorf("ignored table '%s'", r.Table)
	}
	if false {
		if table == TableCollectionView {
			s := string(r.Value)
			fmt.Printf("collection_view json:\n%s\n\n", s)
		}
	}
	if err := json.Unmarshal(r.Value, pRawJSON); err != nil {
		return err
	}
	id := (*pRawJSON)["id"]
	if id != nil {
		r.ID = id.(string)
	}
	if err := json.Unmarshal(r.Value, &obj); err != nil {
		return err
	}
	return nil
}

// LoadPageChunk executes a raw API call /api/v3/loadPageChunk
func (c *webClient) LoadPageChunk(pageID string, chunkNo int, cur *cursor) (*LoadPageChunkResponse, error) { // emulating notion's website api usage: 50 items on first request,
	// 30 on subsequent requests
	limit := 30
	if cur == nil {
		cur = &cursor{
			// to mimic browser api which sends empty array for this argment
			Stack: make([][]stack, 0),
		}
		limit = 50
	}
	req := &loadPageChunkRequest{
		PageID:          pageID,
		ChunkNumber:     chunkNo,
		Limit:           limit,
		Cursor:          *cur,
		VerticalColumns: false,
	}
	var rsp LoadPageChunkResponse
	var err error
	if rsp.RawJSON, err = c.doNotionAPI("/loadPageChunk", req, &rsp); err != nil {
		return nil, err
	}
	for _, r := range rsp.RecordMap.Blocks {
		if err := parseRecord(TableBlock, r); err != nil {
			return nil, err
		}
	}
	return &rsp, nil
}
