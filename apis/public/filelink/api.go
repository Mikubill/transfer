package filelink

import (
	"fmt"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
	"github.com/spf13/cobra"
)

var (
	Backend = new(fileLink)
)

type fileLink struct {
	apis.Backend
	//baseConf *main.MainConfig
	result   string
	Config   cbOptions
	Commands [][]string
}

func (b *fileLink) SetArgs(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf("filelink - https://filelink.io/\n\n" +
		utils.Spacer("  Size Limit: 100M\n") +
		utils.Spacer("  Upload Service: Google Cloud Taiwan/Singapore\n") +
		utils.Spacer("  Download Service: Google Cloud Taiwan/Singapore\n"))
}
