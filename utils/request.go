package utils

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var regex = regexp.MustCompile("filename=\"(.*)\"$")

func DefaultModifier(_ *http.Request) {}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n, err := wc.writer.WriteAt(p, wc.offset)
	if err != nil {
		return 0, err
	}
	wc.offset += int64(n)
	wc.bar.Add(n)
	return n, nil
}

func DownloadFile(filepath, url string, config DownloadConfig) error {

	if url == "" {
		return fmt.Errorf("url is invaild or expired")
	}

	fmt.Printf("fetching download metadata")
	end := DotTicker()

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}
	config.Modifier(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	*end <- struct{}{}

	fmt.Printf("ok\n")

	if resp.StatusCode > 400 {
		return fmt.Errorf("link unavailable, %s", resp.Status)
	}
	length, err := strconv.ParseInt(resp.Header.Get("content-length"), 10, 64)
	if err != nil {
		length = 0
	}
	if IsDir(filepath) {
		if disp := resp.Header.Get("Content-Disposition"); disp != "" {
			dest := regex.FindStringSubmatch(disp)
			if len(dest) > 1 {
				filepath = path.Join(filepath, dest[1])
			} else {
				filepath = path.Join(filepath, GenRandString(4)+".bin")
			}
		} else {
			filepath = path.Join(filepath, GenRandString(4)+".bin")
		}
	}

	fmt.Printf("file save to: %s\n", filepath)

	bar := pb.Full.Start64(0)
	bar.Set(pb.Bytes, true)
	bar.SetTotal(length)

	if IsExist(filepath) && !strings.HasPrefix(filepath, "/dev") {
		return fmt.Errorf("%s exists.(not to overwrite)", filepath)
	}

	fs := syscall.Statfs_t{}
	err = syscall.Statfs(path.Dir(filepath), &fs)
	if err == nil {
		available := int64(fs.Bfree * uint64(fs.Bsize))
		if length > available {
			return fmt.Errorf("no space left on device (path: %s)", path.Dir(filepath))
		}
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	_parallel := 1
	if err := out.Truncate(length); err != nil {
		return fmt.Errorf("tmpfile fruncate failed: %s", err)
	}
	if length > 10*1024*1024 && resp.Header.Get("Accept-Ranges") != "" && config.Parallel > 1 {
		_parallel = config.Parallel
	}

	wg := new(sync.WaitGroup)
	blk := length / int64(_parallel)

	if config.Debug {
		log.Printf("filesize = %d", length)
		log.Printf("parallel = %d", _parallel)
		log.Printf("block = %d", blk)
	}
	for i := 0; i < _parallel; i++ {
		wg.Add(1)
		start := int64(i) * blk
		end := start + blk
		ranger := fmt.Sprintf("%d-%d", start, end)
		if end >= length {
			ranger = fmt.Sprintf("%d-%d", start, length)
		}
		if config.Debug {
			log.Printf("range = %s", ranger)
		}
		counter := &writeCounter{bar: bar, offset: start, writer: out}
		go parallelDownloader(ranger, url, parallelConfig{
			parallel: _parallel,
			modifier: config.Modifier,
			counter:  counter,
			wg:       wg,
		})
	}
	wg.Wait()

	fmt.Print("\n")
	bar.Finish()
	return nil
}

func parallelDownloader(ranger, url string, config parallelConfig) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("createRequest error: %s\n", err)
	}

	req.Header.Set("Range", "bytes="+ranger)
	config.modifier(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("doRequest error: %s\n", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	_, err = io.Copy(ioutil.Discard, io.TeeReader(resp.Body, config.counter))
	if err != nil {
		fmt.Printf("parallel bytes copy returns: %s", err)
	}
	config.wg.Done()
}
