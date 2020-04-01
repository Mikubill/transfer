package main

import (
	"bufio"
	"fmt"
	"os"
	"transfer/apis"
	"transfer/utils"
)

var (
	runConfig = new(utils.MainConfig)
	build     string
)

func init() {
	utils.SetMainArgs(runConfig)
	utils.FlagSet.Usage = func() { utils.PrintUsage(runConfig.Commands) }
	if len(os.Args) >= 2 {
		runConfig.Backend = apis.ParseBackend(os.Args[1])
		if runConfig.Backend != nil {
			runConfig.Backend.SetArgs()
			utils.FlagSet.Usage = func() { utils.PrintUsage(runConfig.Commands, runConfig.Backend.GetArgs()) }
		}
		_ = utils.FlagSet.Parse(os.Args[2:])
	}
}

func main() {

	if len(os.Args) <= 2 || runConfig.Backend == nil {
		utils.FlagSet.Usage()
		return
	}

	files := utils.FlagSet.Args()

	if runConfig.Version {
		printVersion()
		return
	}

	utils.Walker(files, runConfig.Backend)

	if runConfig.KeepMode {
		fmt.Print("Press the enter key to exit...")
		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadString('\n')
	}
}

func printVersion() {
	version := fmt.Sprintf("\ncowTransfer-uploader\n"+
		"Source: https://github.com/Mikubill/cowtransfer-uploader\n"+
		"Build: %s\n", build)
	fmt.Println(version)
}
