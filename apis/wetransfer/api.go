package wetransfer

import (
	"transfer/utils"
)

var (
	Backend = new(weTransfer)
)

type weTransfer struct {
	Config   wssOptions
	Commands [][]string
}

func (b *weTransfer) SetArgs() {
	utils.AddFlag(&b.Config.forceMode, []string{"-force", "-f", "--force"}, false, "Attempt to download file regardless error", &b.Commands)
	utils.AddFlag(&b.Config.debugMode, []string{"-verbose", "-v", "--verbose"}, false, "Verbose Mode", &b.Commands)
	utils.AddFlag(&b.Config.parallel, []string{"-parallel", "-p", "--parallel"}, 4, "Parallel task count (default 4)", &b.Commands)
	utils.AddFlag(&b.Config.interval, []string{"-timeout", "-t", "--timeout"}, 30, "Request retry/timeout limit (in second, default 30)", &b.Commands)
	utils.AddFlag(&b.Config.prefix, []string{"-prefix", "-o", "--output"}, ".", "File download dictionary/name (default \".\")", &b.Commands)
	utils.AddFlag(&b.Config.singleMode, []string{"-single", "-s", "--single"}, false, "Single Upload Mode", &b.Commands)
}

func (b *weTransfer) GetArgs() [][]string {
	return b.Commands
}
