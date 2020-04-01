package chunk

import (
	"bytes"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	cmap "github.com/orcaman/concurrent-map"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"transfer/apis/image"
	"transfer/apis/public/airportal"
	"transfer/utils"
)

func Upload(file string) {
	picBackend := image.ParseBackend(Backend)
	info, err := os.Stat(file)
	if err != nil {
		fmt.Printf("open %s failed: %s", file, err)
		return
	}
	fileStream, err := os.Open(file)
	if err != nil {
		fmt.Printf("open %s failed: %s", file, err)
		return
	}
	ps, _ := filepath.Abs(file)
	fmt.Printf("Local: %s\n", ps)
	bar := pb.Full.Start64(info.Size())
	reader := bar.NewProxyReader(fileStream)
	dataChan := make(chan image.UploadDataFlow)
	for i := 0; i < Parallel; i++ {
		go picBackend.UploadStream(dataChan)
	}
	hashMap := cmap.New()
	wg := new(sync.WaitGroup)
	offset := int64(0)
	for {
		buf := make([]byte, BlockSize)
		nr, err := reader.Read(buf)
		if err == io.EOF {
			if Verbose {
				log.Println(err)
			}
			break
		}
		if err != nil {
			if Verbose {
				log.Println(err)
			}
			fmt.Printf("error reading from connector: %v", err)
		}
		if Verbose {
			log.Println("read", nr, len(buf), offset)
		}
		if nr > 0 {
			//fmt.Println(image.AliBackend.UploadStream(buf[:nr]))
			wg.Add(1)
			item := image.UploadDataFlow{
				Wg:      wg,
				Data:    buf[:nr],
				HashMap: &hashMap,
				Offset:  offset,
			}
			dataChan <- item
		}
		offset += int64(nr)
	}
	wg.Wait()
	bar.Finish()
	close(dataChan)
	fmt.Println("uploading ticket..")
	hashMap.Set("size", strconv.FormatInt(info.Size(), 10))
	hashMap.Set("name", info.Name())
	ticket, err := hashMap.MarshalJSON()
	if err != nil {
		fmt.Println(err)
	}
	backend := airportal.Backend
	filename := utils.GenRandString(6)
	err = backend.PreUpload(filename, int64(len(ticket)))
	if err != nil {
		fmt.Println(err)
	}
	err = backend.DoUpload(filename, int64(len(ticket)), bytes.NewReader(ticket))
	if err != nil {
		fmt.Println(err)
	}
}
