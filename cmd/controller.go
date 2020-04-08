package cmd

import (
	"github.com/spf13/cobra"
	"transfer/apis"
	"transfer/apis/public/airportal"
	"transfer/apis/public/bitsend"
	"transfer/apis/public/catbox"
	"transfer/apis/public/cowtransfer"
	"transfer/apis/public/filelink"
	"transfer/apis/public/gofile"
	"transfer/apis/public/lanzous"
	"transfer/apis/public/tmplink"
	"transfer/apis/public/transfer"
	"transfer/apis/public/vimcn"
	"transfer/apis/public/wenshushu"
	"transfer/apis/public/wetransfer"
)

var (
	baseString = [][]string{
		{"cow", "cowtransfer"},
		{"wss", "wenshushu"},
		{"bit", "bitsend"},
		{"tmp", "tmplink"},
		{"cat", "catbox"},
		{"vim", "vimcn"},
		{"gof", "gofile"},
		{"wet", "wetransfer"},
		{"arp", "airportal"},
		{"flk", "filelink"},
		{"trs", "transfer.sh"},
		{"lzs", "lanzous"},
	}
	baseBackend = []apis.BaseBackend{
		cowtransfer.Backend,
		wenshushu.Backend,
		bitsend.Backend,
		tmplink.Backend,
		catbox.Backend,
		vimcn.Backend,
		gofile.Backend,
		wetransfer.Backend,
		airportal.Backend,
		filelink.Backend,
		transfer.Backend,
		lanzous.Backend,
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
			_ = cmd.Help()
		}
	}
}
