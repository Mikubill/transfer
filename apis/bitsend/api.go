package bitsend

import (
	"transfer/utils"
)

var (
	Backend = new(bitSend)
)

type bitSend struct {
	Config   wssOptions
	Ticket   string
	Commands [][]string
}

func (b *bitSend) SetArgs() {
	utils.AddFlag(&b.Config.forceMode, []string{"-force", "-f", "--force"}, false, "Attempt to download file regardless error", &b.Commands)
	utils.AddFlag(&b.Config.debugMode, []string{"-verbose", "-v", "--verbose"}, false, "Verbose Mode", &b.Commands)
	utils.AddFlag(&b.Config.parallel, []string{"-parallel", "-p", "--parallel"}, 4, "Parallel task count (default 4)", &b.Commands)
	utils.AddFlag(&b.Config.prefix, []string{"-prefix", "-o", "--output"}, ".", "File download dictionary/name (default \".\")", &b.Commands)
	utils.AddFlag(&b.Config.passCode, []string{"-password", "--password"}, "", "Set password", &b.Commands)
}

func (b *bitSend) GetArgs() [][]string {
	return b.Commands
}
