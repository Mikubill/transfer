package anonfiles

import (
	"fmt"

	"github.com/Mikubill/transfer/apis"
	"github.com/spf13/cobra"
)

var (
	Backend = new(anon)
)

type anon struct {
	apis.Backend
	resp     string
	Commands [][]string
}

func (b *anon) SetArgs(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf("anon - https://anonfiles.com/\n\n")
}
