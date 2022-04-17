package musetransfer

import (
	"fmt"
	"sync"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/spf13/cobra"
)

var (
	Backend = new(muse)
)

type muse struct {
	apis.Backend

	Config   museOptions
	EtagMap  *sync.Map
	Assets   []int64
	Commands [][]string
}

func (b *muse) SetArgs(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&b.Config.Parallel, "parallel", "p", 4, "Set the number of upload threads")
	cmd.Flags().IntVarP(&b.Config.interval, "timeout", "t", 30, "Request retry/timeout limit in second")
	cmd.Flags().BoolVarP(&b.Config.singleMode, "single", "s", false, "Single Upload Mode")
	cmd.Long = fmt.Sprintf("Musetransfer - https://musetransfer.com/\n\n" +
		utils.Spacer("  Size Limit: 10G(Anonymous)\n") +
		utils.Spacer("  Upload Service: Aliyun Cloud Storage, Beijing, China\n") +
		utils.Spacer("  Download Service: Aliyun CDN, Global\n"))
}
