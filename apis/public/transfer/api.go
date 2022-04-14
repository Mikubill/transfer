package transfer

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
	"github.com/spf13/cobra"
)

var (
	Backend = new(transfer)
)

type transfer struct {
	apis.Backend
	resp     string
	Commands [][]string
}

func (b *transfer) SetArgs(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf("transfer.sh - https://transfer.sh/\n\n" +
		utils.Spacer("  Size Limit: -\n") +
		utils.Spacer("  Upload Service: Hetzner Online GmbH, Germany\n") +
		utils.Spacer("  Download Service: Hetzner Online GmbH, Germany\n"))
}
