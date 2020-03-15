package pkg

import (
	"fmt"

	"github.com/printzero/tint"
)

var t = tint.Init()

func RubikMsg(message string) {
	template := "\n\n 💠  " + message
	fmt.Println(template)
}

func DebugMsg(message string) {
	template := "@([)@(DEBUG)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.White.Dim(), tint.White.Bold()))
}

func WarnMsg(message string) {
	template := "@([)@(WARN)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.Yellow, tint.White.Bold()))
}

func ErrorMsg(message string) {
	template := "@([)@(ERROR)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.Red, tint.White.Bold()))
}
