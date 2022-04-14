package downloadgg

import (
	"fmt"

	"github.com/Mikubill/transfer/apis"
	"github.com/spf13/cobra"
)

var (
	Backend = new(dlg)
)

type dlg struct {
	apis.Backend
	resp     string
	Commands [][]string
}

func (b *dlg) SetArgs(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf("dlg - https://download.gg/\n\n")
}
