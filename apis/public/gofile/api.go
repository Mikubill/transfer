package gofile

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/spf13/cobra"
)

var (
	Backend = new(goFile)
)

type goFile struct {
	apis.Backend
	totalSize int64

	userToken  string
	serverLink string
	folderID   string
	folderName string

	downloadLink string
	// baseBody     []byte

	// boundary     string
	// streamWriter *io.PipeWriter
	// streamReader *io.PipeReader
	// dataCh       chan []byte

	Config   goFileOptions
	Commands [][]string
}

func (b *goFile) SetArgs(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&b.Config.singleMode, "single", "s", false, "Upload multi files in a single link")
	cmd.Long = fmt.Sprintf("GoFile - https://gofile.io/\n\n" +
		utils.Spacer("  Size Limit: -\n") +
		utils.Spacer("  Upload Service: multi-server, Europe\n") +
		utils.Spacer("  Download Service: multi-server, Europe\n"))
}
