package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Mikubill/transfer/apis"
	fichier "github.com/Mikubill/transfer/apis/public/1fichier"
	"github.com/Mikubill/transfer/apis/public/airportal"
	"github.com/Mikubill/transfer/apis/public/anonfiles"
	"github.com/Mikubill/transfer/apis/public/catbox"
	"github.com/Mikubill/transfer/apis/public/cowtransfer"
	"github.com/Mikubill/transfer/apis/public/downloadgg"
	"github.com/Mikubill/transfer/apis/public/fileio"
	"github.com/Mikubill/transfer/apis/public/gofile"
	"github.com/Mikubill/transfer/apis/public/infura"
	"github.com/Mikubill/transfer/apis/public/lanzous"
	"github.com/Mikubill/transfer/apis/public/litterbox"
	"github.com/Mikubill/transfer/apis/public/musetransfer"
	"github.com/Mikubill/transfer/apis/public/notion"
	"github.com/Mikubill/transfer/apis/public/null"
	"github.com/Mikubill/transfer/apis/public/quickfile"
	"github.com/Mikubill/transfer/apis/public/tmplink"
	"github.com/Mikubill/transfer/apis/public/transfer"
	"github.com/Mikubill/transfer/apis/public/wenshushu"
	"github.com/Mikubill/transfer/apis/public/wetransfer"
)

var (
	backendList = [][]any{
		{"cow", "cowtransfer", cowtransfer.Backend},
		{"wss", "wenshushu", wenshushu.Backend},
		{"tmp", "tmplink", tmplink.Backend},
		{"cat", "catbox", catbox.Backend},
		{"lit", "littlebox", litterbox.Backend},
		{"gof", "gofile", gofile.Backend},
		{"wet", "wetransfer", wetransfer.Backend},
		{"arp", "airportal", airportal.Backend},
		{"trs", "transfer.sh", transfer.Backend},
		{"lzs", "lanzous", lanzous.Backend},
		{"nil", "null", null.Backend},
		{"fio", "file.io", fileio.Backend},
		{"not", "notion", "notion.so", notion.Backend},
		{"fic", "1fichier", fichier.Backend},
		{"inf", "infura", infura.Backend},
		{"muse", "musetransfer", musetransfer.Backend},
		{"qf", "quickfile", quickfile.Backend},
		{"anon", "anonfile", anonfiles.Backend},
		{"gg", "downloadgg", downloadgg.Backend},
	}
)

func ParseLink(link string) apis.BaseBackend {
	for _, item := range backendList {
		backend := item[len(item)-1].(apis.BaseBackend)
		if backend.LinkMatcher(link) {
			return backend
		}
	}
	return nil
}

func inList(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}

func runner(backend apis.BaseBackend) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
		}

		file := uploadWalker(args)
		if len(file) > 0 {
			apis.Upload(file, backend)
		}

		links := downloadWalker(args)
		if len(links) > 0 {
			for _, item := range links {
				backend := ParseLink(item)
				if backend != nil {
					apis.Download(item, backend)
				} else {
					fmt.Println("Unsupported link:", item)
				}
			}
		}

		for k, item := range args {
			isCommand := false
			if strings.HasPrefix(item, "-") {
				isCommand = true
			}
			if k > 1 {
				if strings.HasPrefix(args[k-1], "-") {
					isCommand = true
				}
			}
			if !inList(links, item) && !inList(file, item) && !isCommand {
				fmt.Printf("transfer: %s: No such file, link or directory\n", item)
			}
		}

	}
}
