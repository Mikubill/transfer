package whc

import (
	"fmt"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/spf13/cobra"
)

var (
	Backend = new(whiteCats)
)

type whiteCats struct {
	apis.Backend
	pwd      string
	del      string
	resp     string
	Commands [][]string
}

func (b *whiteCats) SetArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&b.pwd, "password", "p", utils.GenRandString(4), "Set the download password")
	cmd.Flags().StringVarP(&b.del, "delete", "d", utils.GenRandString(4), "Set the remove code")

	cmd.Long = fmt.Sprintf("whiteCat Upload - http://whitecats.dip.jp/\n\n" +
		utils.Spacer("  Size Limit: 100G\n") +
		utils.Spacer("  Upload Service: NTT Communications\n") +
		utils.Spacer("  Download Service: OCN/NTT Communications\n"))
}
