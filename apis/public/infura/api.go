package infura

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/spf13/cobra"
)

var (
	Backend = new(infura)
)

type infura struct {
	apis.Backend
	resp     string
	Commands [][]string
}

func (b *infura) SetArgs(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf("infura - https://infura.io/\n\n" +
		utils.Spacer("  Size Limit: Unspecified \n") +
		utils.Spacer("  Upload Service: AWS, Global\n") +
		utils.Spacer("  Download Service: IPFS, Global\n"))
}
