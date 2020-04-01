package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"transfer/apis"
)

var (
	uploadCmd = &cobra.Command{
		Use:     "upload",
		Aliases: []string{"up"},
		Short:   "Upload a file or dictionary",
		Long: `
Upload a file or dictionary.

Example:
  # specify backend to upload
  transfer upload cow transfer

  # get backend info
  transfer upload cow

`,
	}
)

func init() {

	apis.InitUploadCmd(uploadCmd)
	for n, backend := range baseBackend {
		backendCmd := &cobra.Command{
			Use:     baseString[n][0],
			Aliases: baseString[n],
			Short:   fmt.Sprintf("Use %s API to transfer file", baseString[n][1]),
			Run:     runner(backend),
		}
		backend.SetArgs(backendCmd)
		uploadCmd.AddCommand(backendCmd)
	}
	rootCmd.AddCommand(uploadCmd)
}
