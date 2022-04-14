package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"time"

	"github.com/Mikubill/transfer/utils"
)

var (
	ImgBoxBackend = new(ImgBox)
)

type ImgBox struct {
	picBed
	client http.Client
}

type ImgBoxResp struct {
	Files []struct {
		ID              string    `json:"id"`
		Slug            string    `json:"slug"`
		Name            string    `json:"name"`
		NameHTMLEscaped string    `json:"name_html_escaped"`
		CreatedAt       time.Time `json:"created_at"`
		CreatedAtHuman  string    `json:"created_at_human"`
		UpdatedAt       time.Time `json:"updated_at"`
		GalleryID       any       `json:"gallery_id"`
		URL             string    `json:"url"`
		OriginalURL     string    `json:"original_url"`
		ThumbnailURL    string    `json:"thumbnail_url"`
		SquareURL       string    `json:"square_url"`
		Selected        bool      `json:"selected"`
		CommentsEnabled int       `json:"comments_enabled"`
		CommentsCount   int       `json:"comments_count"`
	} `json:"files"`
}

type ImgBoxTokenResp struct {
	Ok            bool   `json:"ok"`
	TokenID       int    `json:"token_id"`
	TokenSecret   string `json:"token_secret"`
	GalleryID     string `json:"gallery_id"`
	GallerySecret string `json:"gallery_secret"`
}

func (s ImgBox) getToken() (*ImgBoxTokenResp, error) {

	// first get auth_token
	req, err := http.NewRequest("GET", "https://imgbox.com/ajax/token/generate", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	req.Header.Set("referer", "https://imgbox.com/")
	req.Header.Set("accept", "application/json, text/plain, */*")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rep ImgBoxTokenResp
	if err := json.Unmarshal(body, &rep); err != nil {
		return nil, err
	}
	if !rep.Ok {
		return nil, fmt.Errorf("imgbox token error: %s", string(body))
	}

	return &rep, nil
}

func (s ImgBox) Upload(data []byte) (string, error) {
	token, err := s.getToken()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	body, err := s.upload(data, token)
	if err != nil {
		return "", err
	}

	var r ImgBoxResp
	if json.Unmarshal(body, &r); err != nil {
		return "", err
	}

	return r.Files[0].OriginalURL, nil
}

func (s ImgBox) upload(data []byte, token *ImgBoxTokenResp) ([]byte, error) {

	byteBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBuf)
	writer.SetBoundary("------WebKitFormBoundary" + utils.GenRandString(10))

	filename := utils.GenRandString(14) + ".jpg"
	_ = writer.WriteField("token_id", strconv.Itoa(token.TokenID))
	_ = writer.WriteField("token_secret", token.TokenSecret)
	_ = writer.WriteField("content_type", "1")
	_ = writer.WriteField("thumbnail_size", "100c")
	_ = writer.WriteField("comments_enabled", "1")
	_ = writer.WriteField("gallery_id", "null")
	_ = writer.WriteField("gallery_secret", "null")

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="files[]"; filename="%s"`, escapeQuotes(filename)))
	h.Set("Content-Type", "image/jpeg")
	w, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}

	_, _ = w.Write(data)
	_ = writer.Close()

	req, err := http.NewRequest("POST", "https://imgbox.com/upload/process", byteBuf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	// add cookie for upload
	req.Header.Set("referer", "https://imgbox.com/")
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
