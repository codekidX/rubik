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

// DebugMsg appends a debug before the log message
func DebugMsg(message string) {
	template := "@( DEBUG ) " + message
	fmt.Println(t.Exp(template, tint.BgWhite.Add(tint.Black)))
}

// WarnMsg appends a warn before the log message
func WarnMsg(message string) {
	template := "@([)@(WARN)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.Yellow, tint.Normal.Bold()))
}

// ErrorMsg appends a error before the log message
func ErrorMsg(message string) {
	template := "@([)@(ERROR)@(]) " + message
	fmt.Println(t.Exp(template, tint.White.Bold(), tint.Red, tint.Normal.Bold()))
}

// EmojiMsg writes the booting message to stdout
func EmojiMsg(emoji string, message string) {
	template := fmt.Sprintf("  %s  @(%s)", emoji, message)
	fmt.Println(t.Exp(template, tint.Normal.Bold()))
}
