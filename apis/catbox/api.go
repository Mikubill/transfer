package catbox

import (
	"transfer/utils"
)

var (
	Backend = new(catBox)
)

type catBox struct {
	Config   cbOptions
	Commands [][]string
}

func (b *catBox) SetArgs() {
	utils.AddFlag(&b.Config.forceMode, []string{"-force", "-f", "--force"}, false, "Attempt to download file regardless error", &b.Commands)
	utils.AddFlag(&b.Config.debugMode, []string{"-verbose", "-v", "--verbose"}, false, "Verbose Mode", &b.Commands)
	utils.AddFlag(&b.Config.parallel, []string{"-parallel", "-p", "--parallel"}, 4, "Parallel task count (default 4)", &b.Commands)
	utils.AddFlag(&b.Config.prefix, []string{"-prefix", "-o", "--output"}, ".", "File upload dictionary/name (default \".\")", &b.Commands)
}

func (b *catBox) GetArgs() [][]string {
	return b.Commands
}
