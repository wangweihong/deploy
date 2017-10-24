package user

import (
	"os"
	"strings"
)

var (
	UserModuleIP = ""
)

func Init() {
	UserModuleIP = os.Getenv("USERHOST")
	if len(strings.TrimSpace(UserModuleIP)) == 0 {
		panic("must set USERHOST")
	}
}
