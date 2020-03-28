package rubik

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
)

var cmdPrefix = "okrubik> "

// repl is the command to run the REPL for okrubik
func repl() {
	cmd := exec.Command("clear")
	cmd.Run()

	reader := bufio.NewReader(os.Stdin)
	go watchSigint()

	for {
		fmt.Print(cmdPrefix)
		inp, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("An error occured %s. \n okrubik REPL exited. Bye!", err.Error())
		}
		fmt.Println(inp)
	}
}

func watchSigint() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	for range sigChan {
		fmt.Println("\nBye!")
		os.Exit(0)
	}
}
