package apis

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"github.com/Mikubill/transfer/crypto"
	"github.com/Mikubill/transfer/utils"

	"github.com/cheggaaa/pb/v3"
)

var regex = regexp.MustCompile("filename=\"(.*)\"$")

type parallelConfig struct {
	parallel int
	modifier func(r *http.Request)
	counter  *writeCounter
	wg       *sync.WaitGroup
}

type writeCounter struct {
	bar    *pb.ProgressBar
	offset int64
	writer *os.File
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n, err := wc.writer.WriteAt(p, wc.offset)
	if err != nil {
		return 0, err
	}
	wc.offset += int64(n)
	if !NoBarMode && wc.bar != nil {
		wc.bar.Add(n)
	}
	return n, nil
}

func DownloadFile(config *DownloaderConfig) error {
	if Crypto {
		fmt.Println("Warning: crypto mode is enabled. ")
		if config.Config.Parallel != 1 {
			fmt.Println("Note: Crypto mode is not compatible with multi thread download mode, " +
				"setting parallel to 1.")
			config.Config.Parallel = 1
		}
		if Key == "" {
			return fmt.Errorf("crypto mode enabled but encrypt key is not set")
		}
		if len(Key) < 32 {
			Key = string(crypto.Padding([]byte(Key), 32))
		}
		fmt.Printf("Decrypt using key: %s\n", Key)
	}

	if config == nil {
		return nil
	}

	if config.Link == "" {
		return fmt.Errorf("link is invaild or expired\n")
	}

	fmt.Printf("fetching download metadata..")
	end := utils.DotTicker()

	req, err := http.NewRequest("GET", config.Link, nil)
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
		return fmt.Errorf("link unavailable, %s\n", resp.Status)
	}
	if config.RespHandler != nil {
		if !config.RespHandler(resp) {
			return fmt.Errorf("link unavailable.\n")
		}
	}

	length, err := strconv.ParseInt(resp.Header.Get("content-length"), 10, 64)
	if err != nil {
		length = 0
	}

	prefix, err := filepath.Abs(config.Config.Prefix)
	if err != nil {
		return err
	}
	if utils.IsDir(prefix) {
		if config.Config.DebugMode {
			log.Printf("%+v", resp.Header)
		}
		dest := regex.FindStringSubmatch(resp.Header.Get("content-disposition"))
		if len(dest) > 0 {
			exc, err := url.QueryUnescape(strings.TrimSpace(dest[1]))
			if err != nil {
				prefix = path.Join(prefix, strings.TrimSpace(dest[1]))
			} else {
				prefix = path.Join(prefix, exc)
			}
		} else {
			prefix = path.Join(prefix, path.Base(resp.Request.URL.Path))
		}
	}

	fmt.Printf("Saving to:: %s\n", prefix)
	var bar *pb.ProgressBar
	if !NoBarMode {
		bar = pb.Full.Start64(0)
		bar.Set(pb.Bytes, true)
		bar.SetTotal(length)
	}

	if utils.IsExist(prefix) && !strings.HasPrefix(prefix, "/dev") && !config.Config.ForceMode {
		return fmt.Errorf("%s exists.(use -f to overwrite)\n", prefix)
	}

	// not available in windows
	//fs := syscall.Statfs_t{}
	//err = syscall.Statfs(path.Dir(prefix), &fs)
	//if err == nil {
	//	available := int64(fs.Bfree * uint64(fs.Bsize))
	//	if length > available {
	//		return fmt.Errorf("no space left on device (path: %s)", path.Dir(prefix))
	//	}
	//}

	out, err := os.Create(prefix)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	_parallel := 1

	if length > 10*1024*1024 && resp.Header.Get("Accept-Ranges") != "" && config.Config.Parallel > 1 {
		_parallel = config.Config.Parallel
	}

	blk := length / int64(_parallel)

	if config.Config.DebugMode {
		log.Printf("filesize = %d", length)
		log.Printf("parallel = %d", _parallel)
		log.Printf("block = %d", blk)
	}
	if _parallel == 1 {
		if Crypto {
			pipeR, pipeW := io.Pipe()
			blockSize := int64(math.Min(1048576, float64(length)))
			sig := new(sync.WaitGroup)
			sig.Add(1)
			go monitor(pipeW, sig)
			if !NoBarMode && bar != nil {
				go crypto.StreamDecrypt(bar.NewProxyReader(resp.Body), pipeW, Key, blockSize, sig)
			} else {
				go crypto.StreamDecrypt(resp.Body, pipeW, Key, blockSize, sig)
			}
			_, _ = io.Copy(out, pipeR)
		} else {
			if !NoBarMode && bar != nil {
				_, _ = io.Copy(out, bar.NewProxyReader(resp.Body))
			} else {
				_, _ = io.Copy(out, resp.Body)
			}
		}
	} else {
		if err := out.Truncate(length); err != nil {
			return fmt.Errorf("tmpfile fruncate failed: %s\n", err)
		}
		wg := new(sync.WaitGroup)
		for i := 0; i <= _parallel; i++ {
			wg.Add(1)
			start := int64(i) * blk
			end := start + blk
			ranger := fmt.Sprintf("%d-%d", start, end)
			if end >= length {
				ranger = fmt.Sprintf("%d-%d", start, length)
			}
			if config.Config.DebugMode {
				log.Printf("range = %s", ranger)
			}
			counter := &writeCounter{bar: bar, offset: start, writer: out}
			go func() {
				pConf := parallelConfig{
					parallel: _parallel,
					modifier: config.Modifier,
					counter:  counter,
					wg:       wg,
				}
				pRange := ranger
				for {
					err = parallelDownloader(pRange, config.Link, pConf)
					if err == nil {
						break
					}
				}
			}()
		}
		wg.Wait()
	}

	fmt.Print("\n")
	if !NoBarMode && bar != nil {
		bar.Finish()
	}
	return nil
}

func parallelDownloader(ranger, url string, config parallelConfig) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("createRequest error: %s\n", err)
	}

	req.Header.Set("Range", "bytes="+ranger)
	config.modifier(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("doRequest error: %s\n", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	_, err = io.Copy(ioutil.Discard, io.TeeReader(resp.Body, config.counter))
	if err != nil {
		return fmt.Errorf("parallel bytes copy returns: %s", err)
	}
	config.wg.Done()
	return nil
}
