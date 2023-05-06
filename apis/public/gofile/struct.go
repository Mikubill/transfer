package gofile

import "io"

type goFileOptions struct {
	singleMode bool
}

type respBody struct {
	Status string    `json:"status"`
	Data   respBlock `json:"data"`
}

type respBlock struct {
	Server      string              `json:"server"`
	Code        string              `json:"code"`
	RemovalCode string              `json:"adminCode"`
	Items       map[string]fileItem `json:"files"`
}

type fileItem struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Link string `json:"link"`
}

type uploadConfig struct {
	debug      bool
	fileName   string
	fileReader io.Reader
	fileSize   int64
}

type userGen struct {
	Status string `json:"status"`
	Data   struct {
		Token string `json:"token"`
	} `json:"data"`
}

type userDetails struct {
	Status string `json:"status"`
	Data   struct {
		Token      string `json:"token"`
		Email      string `json:"email"`
		Tier       string `json:"tier"`
		RootFolder string `json:"rootFolder"`
	} `json:"data"`
}

type folderDetails struct {
	Status string `json:"status"`
	Data   struct {
		ID           string `json:"id"`
		Type         string `json:"type"`
		Name         string `json:"name"`
		ParentFolder string `json:"parentFolder"`
		CreateTime   int    `json:"createTime"`
		Childs       []any  `json:"childs"`
		Code         string `json:"code"`
	} `json:"data"`
}

// type uploadResult struct {
// 	Status string `json:"status"`
// 	Data   struct {
// 		DownloadPage string `json:"downloadPage"`
// 		Code         string `json:"code"`
// 		ParentFolder string `json:"parentFolder"`
// 		FileID       string `json:"fileId"`
// 		FileName     string `json:"fileName"`
// 		Md5          string `json:"md5"`
// 		DirectLink   string `json:"directLink"`
// 		Info         string `json:"info"`
// 	} `json:"data"`
// }

type sharedDetails struct {
	Status string `json:"status"`
	Data   struct {
	} `json:"data"`
}

type folderDetails2 struct {
	Status string `json:"status"`
	Data   struct {
		IsOwner            bool                   `json:"isOwner"`
		ID                 string                 `json:"id"`
		Type               string                 `json:"type"`
		Name               string                 `json:"name"`
		Childs             []string               `json:"childs"`
		ParentFolder       string                 `json:"parentFolder"`
		Code               string                 `json:"code"`
		CreateTime         int                    `json:"createTime"`
		TotalDownloadCount int                    `json:"totalDownloadCount"`
		TotalSize          int                    `json:"totalSize"`
		Contents           map[string]fileDetails `json:"contents"`
	} `json:"data"`
}

type fileDetails struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	ParentFolder  string `json:"parentFolder"`
	CreateTime    int    `json:"createTime"`
	Size          int    `json:"size"`
	DownloadCount int    `json:"downloadCount"`
	Md5           string `json:"md5"`
	Mimetype      string `json:"mimetype"`
	Viruses       []any  `json:"viruses"`
	ServerChoosen string `json:"serverChoosen"`
	DirectLink    string `json:"directLink"`
	Link          string `json:"link"`
	Thumbnail     string `json:"thumbnail"`
}

type uploadRespData struct {
	DownLoadPage string `json:"downLoadPage"`
	Code         string `json:"code"`
	ParentFolder string `json:"parentFolder"`
	FileId       string `json:"fileId"`
	FileName     string `json:"fileName"`
	Md5          string `json:"md5"`
}
type uploadResp struct {
	Status string         `json:"name"`
	Data   uploadRespData `json:"data"`
}
