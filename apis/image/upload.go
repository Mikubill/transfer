package image

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	Backend string
	Verbose bool
)

func InitCmd(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&Backend,
		"backend", "", "", "Set upload/download backend")
	cmd.Flags().BoolVarP(&Verbose,
		"verbose", "", false, "Enable verbose mode to debug")
}

func Upload(file []string) {
	backend := ParseBackend(Backend)
	for _, v := range file {
		err := filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			data, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Printf("failed: %s", err)
				return nil
			}

			ps, _ := filepath.Abs(path)
			fmt.Printf("Local: %s\n", ps)
			resp, err := backend.Upload(data)
			if err != nil {
				fmt.Printf("failed: %s", err)
				return nil
			}

			fmt.Println(resp)
			return nil
		})
		if err != nil {
			fmt.Printf("filepath.walk(core.upload) returns error: %v, onfile: %s\n", err, v)
		}
	}
}

func ParseBackend(sp string) PicBed {
	switch sp {
	case "ali", "alibaba":
		return AliBackend
	case "bd", "baidu":
		return BDBackend
	case "cc", "ccupload":
		return CCBackend
	case "jj", "juejin":
		return JJBackend
	case "nt", "netease":
		return NTBackend
	case "pr", "prntscr":
		return PRBackend
	case "sm", "smms":
		return SMBackend
	case "sg", "sogou":
		return SGBackend
	case "tt", "toutiao":
		return TTBackend
	case "xm", "xiaomi":
		return XMBackend
	case "vm", "vim", "vimcn":
		return VMBackend
	case "sn", "suning":
		return SNBackend
	default:
		return AliBackend
	}
}
