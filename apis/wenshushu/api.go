package wenshushu

import (
	"transfer/utils"
)

var (
	Backend = new(wssTransfer)
)

type wssTransfer struct {
	Config   wssOptions
	Commands [][]string
}

func (b *wssTransfer) SetArgs() {
	utils.AddFlag(&b.Config.forceMode, []string{"-force", "-f", "--force"}, false, "Attempt to download file regardless error", &b.Commands)
	utils.AddFlag(&b.Config.debugMode, []string{"-verbose", "-v", "--verbose"}, false, "Verbose Mode", &b.Commands)
	utils.AddFlag(&b.Config.token, []string{"-cookie", "-c", "--cookie"}, "", "Your User cookie (optional)", &b.Commands)
	utils.AddFlag(&b.Config.parallel, []string{"-parallel", "-p", "--parallel"}, 4, "Parallel task count (default 4)", &b.Commands)
	utils.AddFlag(&b.Config.blockSize, []string{"-block", "-b", "--block"}, 1048576, "Upload Block Size (default 262144)", &b.Commands)
	utils.AddFlag(&b.Config.interval, []string{"-timeout", "-t", "--timeout"}, 10, "Request retry/timeout limit (in second, default 10)", &b.Commands)
	utils.AddFlag(&b.Config.prefix, []string{"-prefix", "-o", "--output"}, ".", "File download dictionary/name (default \".\")", &b.Commands)
	utils.AddFlag(&b.Config.singleMode, []string{"-single", "-s", "--single"}, false, "Single Upload Mode", &b.Commands)
	utils.AddFlag(&b.Config.passCode, []string{"-password", "--password"}, "", "Set password", &b.Commands)

}

func (b *wssTransfer) GetArgs() [][]string {
	return b.Commands
}
