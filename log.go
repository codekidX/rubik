package cherry

import (
	"fmt"

	"github.com/printzero/tint"
)

var t = tint.Init()

func cherryMsg(message string) {
	template := "@([) 🍒 @(]) " + message
	fmt.Println(t.Exp(template, tint.Green.Bold(), tint.Green.Bold()))
}

func debugMsg(message string) {
	template := "@([)@(DEBUG)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.White.Dim(), tint.White.Bold()))
}

func warnMsg(message string) {
	template := "@([)@(WARN)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.Yellow, tint.White.Bold()))
}

func errorMsg(message string) {
	template := "@([)@(ERROR)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.Red, tint.White.Bold()))
}
