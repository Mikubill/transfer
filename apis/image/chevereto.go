package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Mikubill/transfer/utils"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	CheveretoBackend = new(Chevereto)
)

type Chevereto struct {
	picBed
	dest   string
	client http.Client
}

type CheveretoResp struct {
	StatusCode int `json:"status_code"`
	Image      struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"image"`

	Error struct {
		Message string `json:"message"`
		ErrorID string `json:"errorId"`
		Code    int    `json:"code"`
	} `json:"error"`
	StatusTxt string `json:"status_txt"`
}

func (s Chevereto) getToken() (string, error) {

	// first get auth_token
	req, err := http.NewRequest("GET", s.dest, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	req.Header.Set("referer", s.dest)
	req.Header.Set("accept", "application/json, text/plain, */*")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	tokenRegex := regexp.MustCompile(`auth_token = "([a-f0-9]{40,})";`)
	token := tokenRegex.FindStringSubmatch(string(body))
	if len(token) < 2 {
		return "", fmt.Errorf("token not found")
	}
	result := token[1]

	return result, nil
}

func (s *Chevereto) newUpload(data []byte, dest string) (string, error) {
	if dest == "" {
		return "", fmt.Errorf("chevereto dest is empty")
	}

	if strings.HasPrefix(dest, "http") {
		s.dest = dest
	} else {
		s.dest = "https://" + dest
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}
	s.client = http.Client{Jar: jar, Timeout: 30 * time.Second}

	return s.Upload(data)
}

func (s Chevereto) Upload(data []byte) (string, error) {
	token, err := s.getToken()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	body, err := s.upload(data, token)
	if err != nil {
		return "", err
	}

	var r CheveretoResp
	if json.Unmarshal(body, &r); err != nil {
		return "", err
	}
	if r.StatusCode != 200 {
		return "", fmt.Errorf("%d, %s, %s", r.StatusCode, r.StatusTxt, r.Error.Message)
	}
	return r.Image.URL, nil
}

func (s Chevereto) upload(data []byte, token string) ([]byte, error) {

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	writer.SetBoundary("------WebKitFormBoundary" + utils.GenRandString(10))

	filename := utils.GenRandString(14) + ".jpg"
	_ = writer.WriteField("auth_token", token)
	_ = writer.WriteField("type", "file")
	_ = writer.WriteField("action", "upload")
	_ = writer.WriteField("nsfw", "0")
	_ = writer.WriteField("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="source"; filename="%s"`, escapeQuotes(filename)))
	h.Set("Content-Type", "image/jpeg")
	w, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}

	_, _ = w.Write(data)
	_ = writer.Close()

	APILinkForV3 := fmt.Sprintf("%s/json", s.dest)
	req, err := http.NewRequest("POST", APILinkForV3, byteBuf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	// add cookie for upload
	req.Header.Set("referer", s.dest)
	req.Header.Set("accept", "application/json, text/plain, */*")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%d, %s", resp.StatusCode, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
