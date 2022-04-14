package litterbox

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
	"github.com/spf13/cobra"
)

var (
	Backend = new(litterbox)
)

type litterbox struct {
	apis.Backend
	resp     string
	Commands [][]string
}

func (b *litterbox) SetArgs(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf("litterbox - https://litterbox.catbox.moe/\n\n" +
		utils.Spacer("  Size Limit: 1GB\n") +
		utils.Spacer("  Upload Service: Psychz Networks, Los Angeles California, USA\n") +
		utils.Spacer("  Download Service: Psychz Networks, Los Angeles California, USA\n"))
}
