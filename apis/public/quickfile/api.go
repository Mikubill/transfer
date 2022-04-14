package quickfile

import (
	"fmt"

	"github.com/Mikubill/transfer/apis"
	"github.com/spf13/cobra"
)

var (
	Backend = new(quick)
)

type quick struct {
	apis.Backend
	resp     string
	Commands [][]string
}

func (b *quick) SetArgs(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf("quickfile - https://quickfile.cn/\n\n")
}
