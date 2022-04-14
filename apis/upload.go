package apis

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/Mikubill/transfer/crypto"
	"github.com/Mikubill/transfer/utils"
)

func Upload(files []string, backend BaseBackend) {
	tmpOut := os.Stdout
	if MuteMode {
		transferConfig.NoBarMode = true
		os.Stdout, _ = os.Open(os.DevNull)
	}
	if transferConfig.CryptoMode {
		fmt.Println("Warning: crypto mode is enabled. \n" +
			"Note: Crypto mode still in beta and abnormalities may occur, " +
			"do not over-rely on this function.")
		if transferConfig.CryptoKey == "" || len(transferConfig.CryptoKey) > 32 {
			transferConfig.CryptoKey = utils.GenRandString(16)
			fmt.Printf("Key is not set or incorrect: Setting it to %s\n", transferConfig.CryptoKey)
		}
		if len(transferConfig.CryptoKey) < 32 {
			transferConfig.CryptoKey = string(crypto.Padding([]byte(transferConfig.CryptoKey), 32))
			fmt.Printf("Encrypt using key: %s\n", transferConfig.CryptoKey)
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
			if transferConfig.CryptoMode {
				sizes = append(sizes, crypto.CalcEncryptSize(info.Size()))
			} else {
				sizes = append(sizes, info.Size())
			}
			return nil
		})
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "filepath.walk failed: %v, onfile: %s\n", err, v)
			return
		}
	}
	err := backend.InitUpload(paths, sizes)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error occurred during initialization:\n  %s\n", err)
		return
	}
	for n, file := range paths {
		ps, _ := filepath.Abs(file)
		fmt.Printf("Local: %s\n", ps)
		resp, err := upload(file, sizes[n], backend)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error occurred during upload %s:\n  %s\n", file, err)
		}
		if resp != "" && MuteMode {
			_, _ = fmt.Fprintln(tmpOut, resp)
		}
	}
	resp, err := backend.FinishUpload(files)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error occurred during finalizing upload:\n  %s\n", err)
	}
	if resp != "" && MuteMode {
		_, _ = fmt.Fprintln(tmpOut, resp)
	}
}

func monitor(w *io.PipeWriter, sig *sync.WaitGroup) {
	sig.Wait()
	_ = w.Close()
}

func upload(file string, size int64, backend BaseBackend) (string, error) {
	info, err := os.Stat(file)
	if err != nil {
		return "", fmt.Errorf("stat file %s failed: %s", file, err)
	}
	err = backend.PreUpload(info.Name(), size)
	if err != nil {
		return "", fmt.Errorf("start upload failed: %s", err)
	}
	fileStream, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("open %s failed: %s", file, err)
	}
	var reader io.Reader
	if transferConfig.CryptoMode {
		blockSize := int64(math.Min(1048576, float64(info.Size())))
		pipeR, pipeW := io.Pipe()
		sig := new(sync.WaitGroup)
		sig.Add(1)
		go monitor(pipeW, sig)
		go crypto.StreamEncrypt(fileStream, pipeW, transferConfig.CryptoKey, blockSize, sig)
		reader = pipeR
		if !transferConfig.NoBarMode {
			reader = backend.StartProgress(pipeR, size)
		}
	} else {
		reader = fileStream
		if !transferConfig.NoBarMode {
			reader = backend.StartProgress(fileStream, size)
		}
	}
	err = backend.DoUpload(info.Name(), size, reader)
	if err != nil {
		return "", fmt.Errorf("upload error: %s", err)
	}
	_ = fileStream.Close()
	if !transferConfig.NoBarMode {
		backend.EndProgress()
	}
	resp, err := backend.PostUpload(info.Name(), size)
	if err != nil {
		return "", fmt.Errorf("postUpload error: %s", err)
	}
	return resp, nil
}
