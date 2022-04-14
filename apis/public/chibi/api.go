package chibi

import (
	"fmt"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/spf13/cobra"
)

var (
	Backend = new(chibi)
)

type chibi struct {
	apis.Backend
	resp     string
	Commands [][]string
}

func (b *chibi) SetArgs(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf("Chibisafe/Lolisafe\n\n" +
		utils.Spacer("  Repo: https://github.com/WeebDev/chibisafe\n"))
}
