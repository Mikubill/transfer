package fichier

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/spf13/cobra"
)

var (
	Backend = new(fichier)
)

type fichier struct {
	apis.Backend
	pwd      string
	cookie   string
	resp     string
	remove   string
	Commands [][]string
}

func (b *fichier) SetArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&b.pwd, "password", "p", "", "Set the download password")
	cmd.Flags().StringVarP(&b.cookie, "cookie", "c", "", "Set the user cookie")

	cmd.Long = fmt.Sprintf("1fichier - https://1fichier.com/\n\n" +
		utils.Spacer("  Size Limit: 300G\n") +
		utils.Spacer("  Upload Service: DSTORAGE s.a.s.\n") +
		utils.Spacer("  Download Service: DSTORAGE s.a.s.\n"))
}
