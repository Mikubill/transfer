package musetransfer

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
	"github.com/cheggaaa/pb/v3"
)

const (
	createSend = "https://service.tezign.com/transfer/share/create"
	finishSend = "https://service.tezign.com/transfer/share/finish"
	getSend    = "https://service.tezign.com/transfer/share/get"

	uploadEp   = "https://share-file.tezign.com/"
	downloadEp = "https://musetransfer.com/s/%s"
	getToken   = "https://service.tezign.com/transfer/asset/getUploadToken"
	addFile    = "https://service.tezign.com/transfer/asset/add"

	chunkSize = 1048576 * 2
)

var upIDRegex = regexp.MustCompile(`<UploadId>(\w+)`)

func (b *muse) getUploadToken() (*s3Token, error) {
	req, err := http.NewRequest("GET", getToken, nil)
	if err != nil {
		return nil, err
	}

	addToken(req, b.Config.devicetoken, getToken, []byte(""))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get upload token returns error: %s", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var d getTokenResp
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}
	if d.Message != "success" {
		return nil, fmt.Errorf("get upload token returns error: %s", d.Message)
	}
	return &d.Result, nil
}

func (b *muse) newTransfer() error {
	b.EtagMap = new(sync.Map)
	fmt.Printf("fetching upload tickets..")
	end := utils.DotTicker()

	b.Config.devicetoken = utils.GenRandString(8)[:11]
	if apis.DebugMode {
		log.Println("\ndevice-token: ", b.Config.devicetoken)
	}
	buf, _ := json.Marshal(OrderedMap{
		{"title", utils.GenRandString(10)},
		{"titletype", 0},
		{"expire", 7},
		{"customBackground", 0},
	})

	body, err := b.postAPI(createSend, buf)
	if err != nil {
		return err
	}
	var d createResp
	err = json.Unmarshal(body, &d)
	if err != nil {
		return err
	}
	if d.Message != "success" {
		return fmt.Errorf("create send returns %s", d.Message)
	}
	if d.Result.UploadPath == "" {
		return fmt.Errorf("create send returns empty upload path")
	}

	*end <- struct{}{}
	fmt.Printf("ok\n")
	b.Config.sToken = d.Result.Code
	b.Config.dest = d.Result.UploadPath
	return nil
}

func (b *muse) InitUpload(files []string, sizes []int64) error {
	if b.Config.singleMode {
		b.newTransfer()
	}
	return nil
}

func (b *muse) PreUpload(name string, size int64) error {
	if !b.Config.singleMode {
		b.newTransfer()
	}
	return nil
}

func (b muse) ossSigner(req *http.Request, auth *s3Token, ext ...string) {
	var ctString, md5String string
	if len(ext) > 0 {
		ctString = ext[0]
		md5String = ext[1]
		req.Header["content-md5"] = []string{md5String}
	} else {
		ctString = "application/json"
	}

	gmtdate := time.Now().UTC().Format(http.TimeFormat)
	req.Header["content-type"] = []string{ctString}
	req.Header["x-oss-date"] = []string{gmtdate}
	req.Header["x-oss-user-agent"] = []string{"aliyun-sdk-js/6.17.1 Firefox 99.0 on OS X 10.15"}
	req.Header["x-oss-security-token"] = []string{auth.SecurityToken}
	req.Header["referer"] = []string{"https://musetransfer.com/"}

	params := []string{
		strings.ToUpper(req.Method),
		md5String,
		strings.ToLower(ctString),
		gmtdate,
		fmt.Sprintf("x-oss-date:%s", gmtdate),
		fmt.Sprintf("x-oss-security-token:%s", auth.SecurityToken),
		fmt.Sprintf("x-oss-user-agent:%s", "aliyun-sdk-js/6.17.1 Firefox 99.0 on OS X 10.15"),
	}
	query := strings.TrimSuffix(req.URL.RawQuery, "=")
	params = append(params, fmt.Sprintf("/transfer-private%s?%s", req.URL.Path, query))
	result := strings.Join(params, "\n")

	mac := hmac.New(sha1.New, []byte(auth.AccessKeySecret))
	mac.Write([]byte(result))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	req.Header["Authorization"] = []string{"OSS " + auth.AccessKeyID + ":" + signature}
}

func (b *muse) StartProgress(reader io.Reader, size int64) io.Reader {
	bar := pb.Full.Start64(size)
	bar.Set(pb.Bytes, true)
	b.Bar = bar
	return reader
}

func (b *muse) DoUpload(name string, size int64, file io.Reader) error {

	if apis.DebugMode {
		log.Println("send file init...")
	}
	auth, err := b.getUploadToken()
	if err != nil {
		return err
	}

	link := "https://share-file.tezign.com/" + b.Config.dest + name + "?uploads="

	req, err := http.NewRequest("POST", link, nil)
	if err != nil {
		return err
	}
	b.ossSigner(req, auth)
	if apis.DebugMode {
		log.Println("usq header: ", req.Header)
	}
	http.DefaultClient.Timeout = time.Duration(b.Config.interval) * time.Second
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid response from oss: %s", resp.Status)
	}
	//  using regex to extract ID from response
	id := upIDRegex.FindStringSubmatch(string(body))
	if len(id) != 2 {
		return fmt.Errorf("invalid response: %s", string(body))
	}

	baseURL := "https://share-file.tezign.com/" + b.Config.dest + name + "?partNumber=%d&uploadId=" + id[1]
	// os.Exit(0)

	wg := new(sync.WaitGroup)
	ch := make(chan *uploadPart)
	for i := 0; i < b.Config.Parallel; i++ {
		go b.uploader(&ch)
	}
	part := int64(0)
	for {
		part++
		buf := make([]byte, chunkSize)
		nr, err := io.ReadFull(file, buf)
		if nr <= 0 {
			break
		}
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			fmt.Println(err)
			break
		}
		if nr > 0 {
			wg.Add(1)
			ch <- &uploadPart{
				content: buf[:nr],
				count:   part,
				name:    name,
				wg:      wg,
				auth:    auth,
				dest:    fmt.Sprintf(baseURL, part),
			}
		}
	}

	wg.Wait()
	close(ch)

	link2 := "https://share-file.tezign.com/" + b.Config.dest + name + "?uploadId=" + id[1]
	payload := b.generate_payload(part)
	req, err = http.NewRequest("POST", link2, strings.NewReader(payload))
	if err != nil {
		return err
	}
	b.ossSigner(req, auth)
	http.DefaultClient.Timeout = time.Duration(b.Config.interval) * time.Second
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b.AddEntry(name, b.Config.dest, size)

	return nil
}

func (b *muse) generate_payload(max int64) string {
	var payload string
	payload = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	payload += "<CompleteMultipartUpload>\n"
	for i := int64(1); i <= max; i++ {
		if v, ok := b.EtagMap.LoadAndDelete(i); ok {
			payload += fmt.Sprintf("<Part><PartNumber>%d</PartNumber><ETag>%s</ETag></Part>", i, v)
		}
	}
	payload += "</CompleteMultipartUpload>"
	return payload
}

func (b *muse) AddEntry(n, p string, size int64) error {

	buf, _ := json.Marshal(OrderedMap{
		{"code", b.Config.sToken},
		{"name", n},
		{"path", p + n},
		{"size", size},
		{"type", filepath.Ext(n)},
	})

	body, err := b.postAPI(addFile, buf)

	if err != nil {
		return err
	}

	var d addEntryResp
	err = json.Unmarshal(body, &d)
	if err != nil {
		return err
	}
	if d.Message != "success" {
		return fmt.Errorf("failed to add entry: %s", d.Message)
	}

	b.Assets = append(b.Assets, int64(d.Result.ID))
	return nil
}

func (b muse) PostUpload(string, int64) (string, error) {
	if !b.Config.singleMode {
		return b.completeUpload()
	}
	return "", nil
}

func (b muse) FinishUpload([]string) (string, error) {
	if b.Config.singleMode {
		return b.completeUpload()
	}
	return "", nil
}

func (b muse) completeUpload() (string, error) {

	buf, err := json.Marshal(OrderedMap{
		{"assetIds", b.Assets},
		{"code", b.Config.sToken},
		{"customBackground", 0},
		{"expire", 365},
		{"title", "transfer-" + utils.GenRandString(3)},
		{"titleType", 0},
	})
	if err != nil {
		return "", err
	}

	body, err := b.postAPI(finishSend, buf)
	if err != nil {
		return "", err
	}

	var d finResp
	err = json.Unmarshal(body, &d)
	if err != nil {
		return "", err
	}
	if d.Message != "success" {
		return "", fmt.Errorf(d.Message)
	}

	link := fmt.Sprintf(downloadEp, b.Config.sToken)
	fmt.Printf("Download Link: %s\n", link)
	return link, nil
}

func addToken(req *http.Request, token string, url string, data []byte) {
	addHeaders(req)
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header["x-transfer-device"] = []string{token}
	
	o := base64.StdEncoding.EncodeToString([]byte(strings.TrimPrefix(url, "https://service.tezign.com")))
	i := base64.StdEncoding.EncodeToString(data)
	r := "" //为登录后的token TODO
	a := strings.Join([]string{o, i, token, r}, "|")

	md5Hash := md5.New()
	md5Hash.Write([]byte(a))
	md5Str := hex.EncodeToString(md5Hash.Sum(nil))
	req.Header["x-transfer-sign"] = []string{md5Str}
}

func addHeaders(req *http.Request) {
	req.Header.Set("Referer", "https://musetransfer.com/")
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.82 Safari/537.36")
	req.Header.Set("Origin", "https://musetransfer.com/")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("sec-ch-ua-platform", "Windows")
}
