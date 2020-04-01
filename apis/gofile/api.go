package gofile

import (
	"transfer/utils"
)

var (
	Backend = new(goFile)
)

type goFile struct {
	Config   goFileOptions
	Commands [][]string
}

func (b *goFile) SetArgs() {
	utils.AddFlag(&b.Config.forceMode, []string{"-force", "-f", "--force"}, false, "Attempt to download file regardless error", &b.Commands)
	utils.AddFlag(&b.Config.debugMode, []string{"-verbose", "-v", "--verbose"}, false, "Verbose Mode", &b.Commands)
	utils.AddFlag(&b.Config.parallel, []string{"-parallel", "-p", "--parallel"}, 4, "Parallel task count (default 4)", &b.Commands)
	utils.AddFlag(&b.Config.prefix, []string{"-prefix", "-o", "--output"}, ".", "File upload dictionary/name (default \".\")", &b.Commands)
	utils.AddFlag(&b.Config.token, []string{"-token", "-t", "--token"}, "", "Set your user token (required)", &b.Commands)
}

func (b *goFile) GetArgs() [][]string {
	return b.Commands
}
