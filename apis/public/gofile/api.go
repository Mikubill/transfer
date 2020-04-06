package gofile

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"transfer/apis"
	"transfer/utils"
)

var (
	Backend = new(goFile)
)

type goFile struct {
	apis.Backend
	totalSize int64

	baseBody     []byte
	serverLink   string
	boundary     string
	streamWriter *io.PipeWriter
	streamReader *io.PipeReader
	dataCh       chan []byte

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
