package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

type MainConfig struct {
	Commands    [][]string
	VersionMode bool
	KeepMode    bool
}

var (
	rootCmd = &cobra.Command{
		Use:       "transfer",
		Short:     "Transfer is a very simple big file transfer tool",
		ValidArgs: []string{"image", "tool", "upload", "download"},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if runConfig.VersionMode {
				fmt.Printf("\nTransfer made by Mikubill.\nhttps://github.com/Mikubill/transfer\n\n")
				os.Exit(0)
			}
		},
	}

	runConfig = new(MainConfig)
	//backend apis.BaseBackend
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&runConfig.VersionMode,
		"version", "", false, "Show version and exit")
	rootCmd.PersistentFlags().BoolVarP(&runConfig.KeepMode,
		"keep", "", false, "Keep program active when process finish")
}

func Execute() {

	defer func() {
		if runConfig.KeepMode {
			fmt.Print("Press the enter key to exit...")
			reader := bufio.NewReader(os.Stdin)
			_, _ = reader.ReadString('\n')
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
