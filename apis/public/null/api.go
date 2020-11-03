package null

import (
	"fmt"
	"github.com/spf13/cobra"
	"transfer/apis"
	"transfer/utils"
)

var (
	Backend = new(null)
)

type null struct {
	apis.Backend
	resp     string
	Commands [][]string
}

func (b *null) SetArgs(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf("vim-cn - https://0x0.st/\n\n" +
		utils.Spacer("  Size Limit: 512M\n") +
		utils.Spacer("  Upload Service: Cloudflare, Global\n") +
		utils.Spacer("  Download Service: Cloudflare, Global\n"))
}
