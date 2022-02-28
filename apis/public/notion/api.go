package notion

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/spf13/cobra"
)

var (
	Backend = new(notion)
)

type notion struct {
	apis.Backend
	token    string
	pageID   string
	resp     string
	spaceID  string
	Commands [][]string
}

func (b *notion) SetArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&b.token, "token", "t", "", "Your user cookie (token-v2)")
	cmd.Flags().StringVarP(&b.pageID, "page", "p", "", "Your page id")
	cmd.Flags().StringVarP(&b.spaceID, "space", "s", "", "Your space id")

	cmd.Long = fmt.Sprintf("Notion - https://notion.so/\n\n" +
		utils.Spacer("  Size Limit: 20M(Free), Unlimit(Pro)\n") +
		utils.Spacer("  Upload Service: Amazon S3 US-West\n") +
		utils.Spacer("  Download Service: Amazon S3 US-West, Cloudflare\n"))
}
