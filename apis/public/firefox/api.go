package firefox

import (
	"fmt"

	"github.com/Mikubill/transfer/utils"
)

var (
	Backend = new(ffsend)
)

type ffsend struct {
	Commands [][]string
}

func (b *ffsend) SetArgs() {}

func (b *ffsend) GetInfo() string {
	return fmt.Sprintf("firefox send - https://send.firefox.com/\n\n" +
		utils.Spacer("  Size Limit: 1G(Anonymous), 2.5G(Login)\n"))
}
