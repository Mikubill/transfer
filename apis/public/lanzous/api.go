package lanzous

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/spf13/cobra"
)

var (
	Backend = new(lanzous)
)

type lanzous struct {
	apis.Backend
	resp     string
	Config   wssOptions
	Commands [][]string
}

func (b *lanzous) SetArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&b.Config.token, "cookie", "c", "", "Set your user token (required)")
	cmd.Long = fmt.Sprintf("lanzous - https://www.lanzous.com/\n" +
		"\n  Note: This backend only supports login users. (use -c to set token)\n")
}
