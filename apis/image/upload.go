package image

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	Dest    string
	Backend string
	Verbose bool
)

func InitCmd(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&Backend,
		"backend", "b", "", "Set upload/download backend")
	cmd.Flags().StringVarP(&Dest,
		"dest", "d", "", "Specify domain to upload. (Chevereto)")
	cmd.Flags().BoolVarP(&Verbose,
		"verbose", "v", false, "Enable verbose mode to debug")
}

func Upload(file []string) {
	backend := ParseBackend(Backend)
	for _, v := range file {
		err := filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if err != nil {
				return err
			}

			data, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Printf("failed: %s", err)
				return nil
			}

			ps, _ := filepath.Abs(path)
			fmt.Printf("Local: %s\n", ps)

			var resp string

			if backend != CheveretoBackend {
				resp, err = backend.Upload(data)
			} else {
				if Dest == "" {
					fmt.Println("Error: Chevereto backend need dest domain.")
				}
				resp, err = CheveretoBackend.newUpload(data, Dest)
			}

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
	switch strings.ToLower(sp) {
	// case "ali", "alibaba":
	// 	return AliBackend
	case "bd", "baidu":
		return BDBackend
	case "cc", "ccupload":
		return CCBackend
	// case "jj", "juejin":
	// 	return JJBackend
	// case "nt", "netease":
	// 	return NTBackend
	case "pr", "prntscr":
		return PRBackend
	case "box", "imgbox":
		return ImgBoxBackend
	// case "sm", "smms":
	// 	return SMBackend
	// case "sg", "sogou":
	// 	return SGBackend
	// case "tt", "toutiao":
	// 	return TTBackend
	// case "xm", "xiaomi":
	// 	return XMBackend
	// case "vm", "vim", "vimcn":
	// 	return VMBackend
	// case "sn", "suning":
	// 	return SNBackend
	case "tg", "telegraph":
		return TGBackend
	case "iu", "imgurl":
		return IUBackend
	case "itp", "imgtp":
		return ItpBackend
	case "ikr", "imgkr":
		return IKrBackend
	case "ch", "cheve", "chevereto":
		return CheveretoBackend
	default:
		return CCBackend
	}
}
