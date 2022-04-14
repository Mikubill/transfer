package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/apis/public/fileio"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "transfer",
		Short: "Transfer is a very simple big file transfer tool",
		Long: `
Transfer is a very simple big file transfer tool.

Backend Support:
  airportal(arp), catbox(cat), cowtransfer(cow), fileio(fio),
  gofile(gof), lanzous(lzs), litterbox(lit), null(0x0), transfer(trs),
  wetransfer(wet), vimcn(vim), notion(not), whitecats(whc),
`,
		SilenceErrors: true,
		Example: `  # upload via gofile
  ./transfer gof <your-file>

  # download link
  ./transfer https://.../`,
		Run: func(cmd *cobra.Command, args []string) {
			if VersionMode {
				fmt.Printf("\nTransfer by Mikubill.\nhttps://github.com/Mikubill/transfer\n\n")
				os.Exit(0)
			} else {
				_ = cmd.Help()
			}
		},
	}

	VersionMode bool
	KeepMode    bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&VersionMode,
		"version", "", false, "show version and exit")
	rootCmd.PersistentFlags().BoolVarP(&KeepMode,
		"keep", "", false, "keep program active when process finish")
	apis.InitCmd(rootCmd)
	for _, item := range backendList {
		backend := item[len(item)-1].(apis.BaseBackend)
		var alias []string
		for _, a := range item {
			if _, ok := a.(string); ok {
				alias = append(alias, a.(string))
			}
		}
		backendCmd := &cobra.Command{
			Use:     alias[0],
			Aliases: alias[1:],
			Short:   fmt.Sprintf("Use %s API to transfer file", alias[1]),
			Run:     runner(backend),
		}
		backend.SetArgs(backendCmd)
		backendCmd.Hidden = true
		rootCmd.AddCommand(backendCmd)
	}
}

func Execute() {

	defer func() {
		if KeepMode {
			fmt.Print("Press the enter key to exit...")
			reader := bufio.NewReader(os.Stdin)
			_, _ = reader.ReadString('\n')
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		if strings.HasPrefix(err.Error(), "unknown command") {
			handleRootTransfer(os.Args[1:])
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func handleRootTransfer(args []string) {

	rootCmd.ParseFlags(args)
	files := uploadWalker(args)
	if len(files) != 0 {
		if !apis.MuteMode {
			fmt.Println("Warning: backend is not set. Using default: fileio.backend - <file.io>")
			fmt.Printf("Run 'transfer --help' for usage.\n\n")
		}
	}
	runner(fileio.Backend)(rootCmd, args)
}
