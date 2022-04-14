package crypto

import (
	"fmt"
	"github.com/Mikubill/transfer/utils"
	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ForceMode bool
	Prefix    string
	Key       string
	NoBar     bool
	blockSize int64
)

func InitCmd(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&Prefix,
		"output", "o", ".", "Write to another file/folder")
	cmd.Flags().StringVarP(&Key,
		"key", "k", "", "Set encrypt/decrypt password")
	cmd.Flags().BoolVarP(&ForceMode,
		"force", "f", false, "Attempt to process file regardless error")
	cmd.Flags().BoolVarP(&NoBar,
		"no-progress", "", false, "Disable Progress Bar to reduce output")
}

func Encrypt(file string) error {
	path, err := filepath.Abs(file)
	if err != nil {
		return err
	}

	var dest string
	if utils.IsDir(Prefix) {
		dest = filepath.Join(Prefix, filepath.Base(file)+".encrypt")
	} else {
		dest = Prefix
	}

	dest, err = filepath.Abs(dest)
	if err != nil {
		return err
	}

	if utils.IsExist(dest) && !strings.HasPrefix(dest, "/dev") && !ForceMode {
		return fmt.Errorf("%s exists.(use -f to overwrite)", dest)
	}

	fmt.Printf("Local: %s\nDest: %s\n", path, dest)

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	blockSize = int64(math.Min(1048576, float64(info.Size())))

	src, err := os.Open(path)
	if err != nil {
		return err
	}

	enc, err := os.Create(dest)
	if err != nil {
		return err
	}

	var writer io.Writer
	var bar *pb.ProgressBar
	if !NoBar {
		bar = pb.Full.Start64(CalcEncryptSize(info.Size()))
		writer = bar.NewProxyWriter(enc)
	} else {
		writer = enc
	}

	if Key == "" || len(Key) > 32 {
		Key = utils.GenRandString(16)
		fmt.Printf("Key is not set or incorrect: Setting it to %s\n", Key)
	}

	if len(Key) < 32 {
		Key = string(Padding([]byte(Key), 32))
		//fmt.Printf("Key is not set or incorrect: Setting it to %s\n", Key)
	}

	sig := new(sync.WaitGroup)
	sig.Add(1)
	StreamEncrypt(src, writer, Key, blockSize, sig)
	if !NoBar && bar != nil {
		bar.Finish()
	}
	_ = src.Close()
	_ = enc.Close()
	return nil
}

func Decrypt(file string) error {
	path, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	var dest string

	if utils.IsDir(Prefix) {
		dest = filepath.Join(Prefix, strings.Replace(file, ".encrypt", "", 1))
	} else {
		dest = Prefix
	}

	dest, err = filepath.Abs(dest)
	if err != nil {
		return err
	}

	if utils.IsExist(dest) && !strings.HasPrefix(dest, "/dev") && !ForceMode {
		return fmt.Errorf("%s exists.(use -f to overwrite)", dest)
	}

	fmt.Printf("Local: %s\nDest: %s\n", path, dest)

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	blockSize = int64(math.Min(1048576, float64(info.Size())))

	src, err := os.Open(path)
	if err != nil {
		return err
	}

	dec, err := os.Create(dest)
	if err != nil {
		return err
	}

	var writer io.Writer
	var bar *pb.ProgressBar
	if !NoBar {
		bar = pb.Full.Start64(info.Size())
		writer = bar.NewProxyWriter(dec)
	} else {
		writer = dec
	}

	if Key == "" || len(Key) > 32 {
		return fmt.Errorf("key is not set")
	}

	if len(Key) < 32 {
		Key = string(Padding([]byte(Key), 32))
		fmt.Printf("Key is not set or incorrect: Setting it to %s\n", Key)
	}

	sig := new(sync.WaitGroup)
	sig.Add(1)
	StreamDecrypt(src, writer, Key, blockSize, sig)
	if !NoBar && bar != nil {
		bar.Finish()
	}
	_ = src.Close()
	_ = dec.Close()
	return nil
}
