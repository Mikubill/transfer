package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"transfer/apis"
	"transfer/apis/public/filelink"
)

var (
	rootCmd = &cobra.Command{
		Use:   "transfer",
		Short: "Transfer is a very simple big file transfer tool",
		Long: `
Transfer is a very simple big file transfer tool.

Backend Support:
  arp  -  Airportal  -  https://aitportal.cn/
  bit  -  bitSend  -  https://bitsend.jp/
  cat  -  CatBox  -  https://catbox.moe/    
  cow  -  CowTransfer  -  https://www.cowtransfer.com/  
  gof  -  GoFile  -  https://gofile.io/
  tmp  -  TmpLink  -  https://tmp.link/     
  vim  -  Vim-cn  -  https://img.vim-cn.com/    
  wss  -  WenShuShu  -  https://www.wenshushu.cn/  
  wet  -  WeTransfer  -  https://wetransfer.com/  
  flk  -  FileLink  -  https://filelink.io/
  trs  -  Transfer.sh  -  https://transfer.sh/
  lzs  -  Lanzous  -  https://www.lanzous.com/
`,
		SilenceErrors: true,
		Example: `  # upload via wenshushu
  ./transfer wss <your-file>

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
	for n, backend := range baseBackend {
		backendCmd := &cobra.Command{
			Use:     baseString[n][0],
			Aliases: baseString[n],
			Short:   fmt.Sprintf("Use %s API to transfer file", baseString[n][1]),
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

	_ = rootCmd.ParseFlags(args)
	links := downloadWalker(args)
	if len(links) != 0 {
		for _, item := range links {
			backend := ParseLink(item)
			if backend != nil {
				apis.Download(item, backend)
			}
		}
		return
	}

	files := uploadWalker(args)
	if len(files) != 0 {
		fmt.Println("Warning: backend is not set. Using default: filelink.backend - <filelink.io>")
		fmt.Printf("Run 'transfer --help' for usage.\n\n")
		apis.Upload(files, filelink.Backend)
		return
	}

	fmt.Println("Error: no file/url detected.")
	fmt.Println("Use \"transfer --help\" for more information.")
}
