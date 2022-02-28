package fileio

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/spf13/cobra"
)

var (
	Backend = new(fileio)
)

type fileio struct {
	apis.Backend
	resp     string
	Commands [][]string
}

func (b *fileio) SetArgs(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf("file-io - https://file.io/\n\n" +
		utils.Spacer("  Size Limit: 100M\n") +
		utils.Spacer("  Upload Service: Cloudflare, Global\n") +
		utils.Spacer("  Download Service: Cloudflare, Global\n"))
}
