package mode

import (
	"github.com/astaxie/beego"
)

const (
	ProductMode ModeType = "product"
	DevelopMode ModeType = "develop"
)

type ModeType string

var (
	CurrentMode ModeType
)

func TurnProductModeOn() {

	CurrentMode = ProductMode
}

func TurnProductModeOff() {
	CurrentMode = DevelopMode

}
func IsProductMode() bool {
	return CurrentMode == ProductMode
}
func IsDevelopMode() bool {
	return CurrentMode == DevelopMode
}

func init() {
	if beego.BConfig.RunMode == "dev" {
		CurrentMode = DevelopMode
	} else {
		CurrentMode = ProductMode
	}
}
