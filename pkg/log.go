package pkg

import (
	"fmt"

	"github.com/printzero/tint"
)

var t = tint.Init()

// Logger is the go to logging struct for anything related to logs
type Logger struct {
	CanLog bool
}

// RubikMsg appends a diamond before the log message
func RubikMsg(message string) {
	template := "\n\n 💠  " + message
	fmt.Println(template)
}

// DebugMsg appends a debug before the log message
func DebugMsg(message string) {
	template := "@([)@(DEBUG)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.White.Dim(), tint.White.Bold()))
}

// WarnMsg appends a warn before the log message
func WarnMsg(message string) {
	template := "@([)@(WARN)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.Yellow, tint.White.Bold()))
}

// ErrorMsg appends a error before the log message
func ErrorMsg(message string) {
	template := "@([)@(ERROR)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.Red, tint.White.Bold()))
}
