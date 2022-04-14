package bitsend

import (
	"fmt"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
	"github.com/spf13/cobra"
)

var (
	Backend = new(bitSend)
)

type bitSend struct {
	apis.Backend
	resp     uploadResp
	Config   wssOptions
	Ticket   string
	Commands [][]string
}

func (b *bitSend) SetArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&b.Config.passCode, "password", "p", "", "Set upload password")
	cmd.Long = fmt.Sprintf("BitSend - https://bitsend.jp/\n\n" +
		utils.Spacer("  Size Limit: -\n") +
		utils.Spacer("  Upload Service: OVH SAS, Boa Nou Quebec, Canada\n") +
		utils.Spacer("  Download Service: OVH SAS, Boa Nou Quebec, Canada\n"))
}
