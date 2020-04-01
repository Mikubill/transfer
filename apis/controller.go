package apis

import (
	"transfer/apis/bitsend"
	"transfer/apis/catbox"
	"transfer/apis/cowtransfer"
	"transfer/apis/gofile"
	"transfer/apis/tmplink"
	"transfer/apis/vimcn"
	"transfer/apis/wenshushu"
	"transfer/apis/wetransfer"
	"transfer/utils"
)

func ParseBackend(name string) utils.Backend {
	switch name {
	case "cow", "cowtransfer", "c-t":
		return cowtransfer.Backend
	case "wss", "wsstransfer", "wenshushu":
		return wenshushu.Backend
	case "bit", "bitsend":
		return bitsend.Backend
	case "tmp", "tmplink":
		return tmplink.Backend
	case "cat", "catbox":
		return catbox.Backend
	case "vim", "vimcn":
		return vimcn.Backend
	case "gof", "gofile":
		return gofile.Backend
	case "wet", "wetransfer":
		return wetransfer.Backend
	}
	return nil
}
