package chunk

import (
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"transfer/apis/image"
	"transfer/apis/public/airportal"
	"transfer/utils"
)

var Matcher = airportal.Backend.LinkMatcher

func Download(file string) {
	picBackend := image.ParseBackend(Backend)
	dataChan := make(chan image.DownloadDataFlow)
	for i := 0; i < Parallel; i++ {
		go picBackend.DownloadStream(dataChan)
	}
	req, err := http.NewRequest("GET", file, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 Transfer/0.1.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	bd, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	if Verbose {
		log.Println(string(bd))
	}
	_ = resp.Body.Close()
	var ticket map[string]string
	if err := json.Unmarshal(bd, &ticket); err != nil {
		fmt.Println(err)
		return
	}
	wg := new(sync.WaitGroup)
	prefix, err := filepath.Abs(Prefix)
	if err != nil {
		fmt.Printf("prefix is invalid: %v", err)
		return
	}
	if utils.IsDir(prefix) {
		prefix = path.Join(prefix, ticket["name"])
	}
	if utils.IsExist(prefix) && !strings.HasPrefix(prefix, "/dev") && !ForceMode {
		fmt.Printf("%s exists.(use -f to overwrite)", prefix)
		return
	}
	out, err := os.Create(prefix)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = out.Close()
	}()
	length, err := strconv.ParseInt(ticket["size"], 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := out.Truncate(length); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Save file to: " + ticket["name"])
	bar := pb.Full.Start64(length)
	bar.Set(pb.Bytes, true)
	for k, v := range ticket {
		if k != "size" && k != "name" {
			if Verbose {
				log.Println(k, v)
			}
			wg.Add(1)
			item := image.DownloadDataFlow{
				Wg:     wg,
				File:   out,
				Hash:   v,
				Offset: k,
				Bar:    bar,
			}
			dataChan <- item
		}
	}
	wg.Wait()
	close(dataChan)
	bar.Finish()
	_ = out.Close()
}
