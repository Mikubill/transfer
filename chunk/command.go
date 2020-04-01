package chunk

import "github.com/spf13/cobra"

var (
	Verbose   bool
	Prefix    string
	ForceMode bool
	BlockSize int
	Parallel  int
	Backend   string
)

func InitCmd(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&Prefix,
		"output", "o", ".", "Write to another file/dictionary")
	cmd.Flags().BoolVarP(&Verbose,
		"verbose", "v", false, "Verbose mode to debug")
	cmd.Flags().BoolVarP(&ForceMode,
		"force", "f", false, "Attempt to download file regardless error")
	cmd.Flags().IntVarP(&BlockSize,
		"block", "b", 1048576, "Set upload chunk size")
	cmd.Flags().IntVarP(&Parallel,
		"parallel", "p", 3, "Set upload/download task count")
	cmd.Flags().StringVarP(&Backend,
		"backend", "", "", "Set upload/download backend")
}
