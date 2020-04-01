package apis

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"
	"transfer/crypto"
	"transfer/utils"
)

func Upload(files []string, backend BaseBackend) {
	if Crypto {
		fmt.Println("Warning: crypto mode is enabled. \n" +
			"Note: Crypto mode is still in beta and abnormalities may occur, " +
			"do not over-rely on this function.")
		if Key == "" || len(Key) > 32 {
			Key = utils.GenRandString(16)
			fmt.Printf("Key is not set or incorrect: Setting it to %s\n", Key)
		}
		if len(Key) < 32 {
			Key = string(crypto.Padding([]byte(Key), 32))
			fmt.Printf("Encrypt using key: %s\n", Key)
		}

	}
	var (
		sizes []int64
		paths []string
	)
	for _, v := range files {
		err := filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if err != nil {
				return err
			}
			paths = append(paths, path)
			if Crypto {
				sizes = append(sizes, crypto.CalcEncryptSize(info.Size()))
			} else {
				sizes = append(sizes, info.Size())
			}
			return nil
		})
		if err != nil {
			fmt.Printf("filepath.walk failed: %v, onfile: %s\n", err, v)
			return
		}
	}
	err := backend.InitUpload(paths, sizes)
	if err != nil {
		fmt.Printf("init Upload Error: %s\n", err)
		return
	}
	for n, file := range paths {
		err := upload(file, sizes[n], backend)
		if err != nil {
			fmt.Printf("upload %s failed: %s\n", file, err)
		}
	}
	err = backend.FinishUpload(files)
	if err != nil {
		fmt.Printf("Finish Upload Error: %s\n", err)
	}
}

func monitor(w *io.PipeWriter, sig *sync.WaitGroup) {
	sig.Wait()
	_ = w.Close()
}

func upload(file string, size int64, backend BaseBackend) error {
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("stat file %s failed: %s\n", file, err)
	}
	err = backend.PreUpload(info.Name(), size)
	if err != nil {
		return fmt.Errorf("start upload failed: %s", err)
	}
	fileStream, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open %s failed: %s", file, err)
	}
	ps, _ := filepath.Abs(file)
	fmt.Printf("Local: %s\n", ps)
	var reader io.Reader
	bar := pb.Full.Start64(size)
	if Crypto {
		blockSize := int64(math.Min(1048576, float64(info.Size())))
		pipeR, pipeW := io.Pipe()
		sig := new(sync.WaitGroup)
		sig.Add(1)
		go monitor(pipeW, sig)
		go crypto.StreamEncrypt(fileStream, pipeW, Key, blockSize, sig)
		reader = bar.NewProxyReader(pipeR)
	} else {
		bar = pb.Full.Start64(size)
		reader = bar.NewProxyReader(fileStream)
	}
	err = backend.DoUpload(info.Name(), size, reader)
	if err != nil {
		return fmt.Errorf("Do Upload Error: %s\n", err)
	}
	bar.Finish()
	_ = fileStream.Close()
	err = backend.PostUpload(info.Name(), size)
	if err != nil {
		return fmt.Errorf("PostUpload Error: %s\n", err)
	}
	return nil
}
