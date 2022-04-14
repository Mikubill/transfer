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
	if wc.bar != nil {
		wc.bar.Add(n)
	}
	return n, nil
}

func DownloadFile(config DownConfig) error {
	if config.CryptoMode {
		fmt.Println("Warning: crypto mode is enabled. ")
		if config.Parallel != 1 {
			fmt.Println("Note: Crypto mode is not compatible with multi thread download mode, " +
				"setting parallel to 1.")
			config.Parallel = 1
		}
		if config.CryptoKey == "" {
			return fmt.Errorf("crypto mode enabled but encrypt key is not set")
		}
		if len(config.CryptoKey) < 32 {
			config.CryptoKey = string(crypto.Padding([]byte(config.CryptoKey), 32))
		}
		fmt.Printf("Decrypt using key: %s\n", config.CryptoKey)
	}

	if config.Link == "" {
		return fmt.Errorf("link is invaild or expired")
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
		return fmt.Errorf("link unavailable, %s", resp.Status)
	}
	if config.RespHandler != nil {
		if !config.RespHandler(resp) {
			return fmt.Errorf("link unavailable.")
		}
	}

	length, err := strconv.ParseInt(resp.Header.Get("content-length"), 10, 64)
	if err != nil {
		length = 0
	}

	prefix, err := filepath.Abs(config.Prefix)
	if err != nil {
		return err
	}
	if utils.IsDir(prefix) {
		if DebugMode {
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
	if !config.NoBarMode {
		bar = pb.Full.Start64(0)
		bar.Set(pb.Bytes, true)
		bar.SetTotal(length)
	}

	if utils.IsExist(prefix) && !strings.HasPrefix(prefix, "/dev") && !config.ForceMode {
		return fmt.Errorf("%s exists.(use -f to overwrite)", prefix)
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

	if length > 10*1024*1024 && resp.Header.Get("Accept-Ranges") != "" && config.Parallel > 1 {
		_parallel = config.Parallel
	}

	blk := length / int64(_parallel)

	if DebugMode {
		log.Printf("filesize = %d", length)
		log.Printf("parallel = %d", _parallel)
		log.Printf("block = %d", blk)
	}
	if _parallel == 1 {
		if config.CryptoMode {
			pipeR, pipeW := io.Pipe()
			blockSize := int64(math.Min(1048576, float64(length)))
			sig := new(sync.WaitGroup)
			sig.Add(1)
			go monitor(pipeW, sig)
			if !config.NoBarMode && bar != nil {
				go crypto.StreamDecrypt(bar.NewProxyReader(resp.Body), pipeW, config.CryptoKey, blockSize, sig)
			} else {
				go crypto.StreamDecrypt(resp.Body, pipeW, config.CryptoKey, blockSize, sig)
			}
			_, _ = io.Copy(out, pipeR)
		} else {
			if !config.NoBarMode && bar != nil {
				_, _ = io.Copy(out, bar.NewProxyReader(resp.Body))
			} else {
				_, _ = io.Copy(out, resp.Body)
			}
		}
	} else {
		if err := out.Truncate(length); err != nil {
			return fmt.Errorf("tmpfile fruncate failed: %s", err)
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
			if DebugMode {
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
	if bar != nil {
		bar.Finish()
	}
	return nil
}

func parallelDownloader(ranger, url string, config parallelConfig) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("createRequest error: %s", err)
	}

	req.Header.Set("Range", "bytes="+ranger)
	config.modifier(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("doRequest error: %s", err)
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
