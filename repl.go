package rubik

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/rubikorg/rubik/pkg"
	"github.com/rubikorg/rubik/replc"
)

var replPrefix = "okrubik> "

var projectPath = ""

var cmdMap = map[string]func([]string) string{
	"select": selectCommand,
	"exit":   exitCommand,
	"help":   replc.HelpCommand,
	"list":   replc.ListCommand,
}

// because we need to keep the project information on the root
// we defined the selectCommand here. This saves the project path
// used for evaluation
func selectCommand(args []string) string {
	config, err := pkg.GetRubikConfig()
	if err != nil {
		return err.Error()
	}

	var name string
	if len(args) > 1 {
		name = args[1]
	} else {
		return "Select command needs argument with app name"
	}

	// we will always get a rubik config because REPL will open
	// only when there is rubik.toml in the path
	for _, p := range config.App {
		if p.Name == name {
			projectPath = p.Path
			break
		}
	}

	if projectPath == "" {
		return "Cannot find your app in rubik.toml file"
	}

	pwd, _ := os.Getwd()
	if !strings.Contains(projectPath, "./") {
		return "Bad rubik.toml. The path to a project must be a relative path. Example: " +
			"./cmd/server"
	}

	projectPath = strings.ReplaceAll(projectPath, "./", pwd)

	return fmt.Sprintf("Selected project: %s", name)
}

func exitCommand(args []string) string {
	fmt.Println("Bye!")
	os.Exit(0)
	return ""
}

// repl is the command to run the REPL for okrubik
func repl() {
	cmd := exec.Command("clear")
	cmd.Run()

	reader := bufio.NewReader(os.Stdin)
	go watchSigint()

	for {
		fmt.Print(replPrefix)
		inp, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("An error occured %s. \n okrubik REPL exited. Bye!", err.Error())
		}
		trinp := strings.TrimSpace(inp)

		cmd := strings.Fields(trinp)
		fn := cmdMap[cmd[0]]
		if fn == nil {
			fmt.Println("No such command")
		} else {
			fmt.Println(fn(cmd))
		}
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
