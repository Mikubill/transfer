package cmd

import (
	"github.com/Mikubill/transfer/apis"
	fichier "github.com/Mikubill/transfer/apis/public/1fichier"
	"github.com/Mikubill/transfer/apis/public/airportal"
	"github.com/Mikubill/transfer/apis/public/fileio"
	"github.com/Mikubill/transfer/apis/public/infura"
	"github.com/Mikubill/transfer/apis/public/null"

	"github.com/spf13/cobra"

	//"transfer/apis/public/bitsend"
	"github.com/Mikubill/transfer/apis/public/catbox"
	"github.com/Mikubill/transfer/apis/public/cowtransfer"

	// "transfer/apis/public/filelink"
	"github.com/Mikubill/transfer/apis/public/gofile"
	"github.com/Mikubill/transfer/apis/public/lanzous"
	"github.com/Mikubill/transfer/apis/public/litterbox"

	//"transfer/apis/public/tmplink"
	"github.com/Mikubill/transfer/apis/public/transfer"
	// "transfer/apis/public/vimcn"

	"github.com/Mikubill/transfer/apis/public/notion"
	"github.com/Mikubill/transfer/apis/public/wenshushu"
	"github.com/Mikubill/transfer/apis/public/wetransfer"
	// whc "transfer/apis/public/whitecats"
)

var (
	baseString = [][]string{
		{"cow", "cowtransfer"},
		{"wss", "wenshushu"},
		// {"bit", "bitsend"},
		// {"tmp", "tmplink"},
		{"cat", "catbox"},
		{"lit", "littlebox"},
		// {"vim", "vimcn"},
		{"gof", "gofile"},
		{"wet", "wetransfer"},
		{"arp", "airportal"},
		// {"flk", "filelink"},
		{"trs", "transfer.sh"},
		{"lzs", "lanzous"},
		{"0x0", "null"},
		{"fio", "file.io"},
		{"not", "notion", "notion.so"},
		// {"whc", "whitecat"},
		{"fic", "1fichier"},
		{"inf", "infura"},
	}
	baseBackend = []apis.BaseBackend{
		cowtransfer.Backend,
		wenshushu.Backend,
		//bitsend.Backend,
		//tmplink.Backend,
		catbox.Backend,
		litterbox.Backend,
		// vimcn.Backend,
		gofile.Backend,
		wetransfer.Backend,
		airportal.Backend,
		// filelink.Backend,
		transfer.Backend,
		lanzous.Backend,
		null.Backend,
		fileio.Backend,
		notion.Backend,
		// whc.Backend,
		fichier.Backend,
		infura.Backend,
	}
)

func ParseLink(link string) apis.BaseBackend {
	for _, item := range baseBackend {
		if item.LinkMatcher(link) {
			return item
		}
	}
	return nil
}

func runner(backend apis.BaseBackend) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		file := uploadWalker(args)
		if len(file) != 0 {
			apis.Upload(file, backend)
		} else {
			links := downloadWalker(args)
			if len(links) != 0 {
				for _, item := range links {
					backend := ParseLink(item)
					if backend != nil {
						apis.Download(item, backend)
					}
				}
				return
			} else {
				_ = cmd.Help()
			}
		}
	}
}
