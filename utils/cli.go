package utils

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unsafe"
)

var FlagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func SetMainArgs(b *MainConfig) {
	AddFlag(&b.KeepMode, []string{"-keep", "-k", "--keep"}, false, "Keep program active when upload/download finish", &b.Commands)
}

func PrintUsage(main [][]string, backend ...[][]string) {
	if len(backend) == 0 {
		fmt.Printf("\nUsage:\n\n  %s <backend> [options] file(s)/url(s)\n\n"+
			"Backend Support:\n\n"+
			"  cow - Cowtransfer https://www.cowtransfer.com/\n"+
			"  wss - Wenshushu https://www.wenshushu.cn/\n"+
			"  bit - BitSend https://www.bitsend.jp/\n"+
			"  tmp - TmpLink https://tmp.link/\n"+
			"  cat - CatBox https://catbox.moe/\n"+
			"  vim - Vim-CN https://img.vim-cn.com/\n"+
			"  gof - GoFile https://gofile.io/\n"+
			"  wet - WeTransfer https://wetransfer.com/\n\n", os.Args[0])
	} else {
		fmt.Printf("\nUsage:\n\n  %s %s [options] file(s)/url(s)\n\n", os.Args[0], os.Args[1])
	}
	fmt.Printf("Global Options:\n\n")
	for _, val := range main {
		s := fmt.Sprintf("  %s %s", val[0], val[1])
		block := strings.Repeat(" ", 30-len(s))
		fmt.Printf("%s%s%s\n", s, block, val[2])
	}
	fmt.Printf("\n")

	if len(backend) > 0 {
		fmt.Printf("Backend Options:\n\n")
		for _, val := range backend[0] {
			s := fmt.Sprintf("  %s %s", val[0], val[1])
			block := strings.Repeat(" ", 30-len(s))
			fmt.Printf("%s%s%s\n", s, block, val[2])
		}
		fmt.Printf("\n")
	} else {
		return
	}
}

func AddFlag(p interface{}, cmd []string, val interface{}, usage string, cli *[][]string) {
	s := []string{strings.Join(cmd[1:], ", "), "", usage}
	ptr := unsafe.Pointer(reflect.ValueOf(p).Pointer())
	for _, item := range cmd {
		switch val := val.(type) {
		case int:
			s[1] = "int"
			FlagSet.IntVar((*int)(ptr), item[1:], val, usage)
		case string:
			s[1] = "string"
			FlagSet.StringVar((*string)(ptr), item[1:], val, usage)
		case bool:
			FlagSet.BoolVar((*bool)(ptr), item[1:], val, usage)
		}
	}
	*cli = append(*cli, s)
}
