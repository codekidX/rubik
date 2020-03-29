package replc

import (
	"fmt"
)

type cmd struct {
	name  string
	desc  string
	usage string
}

var allCommands = []cmd{
	cmd{
		name:  "create",
		desc:  "creates a new rubik project",
		usage: "okrubik create",
	},
}

var replCommands = []cmd{
	cmd{
		name:  "select",
		desc:  "selects the project that you have initialized inside rubik.toml",
		usage: "select [project_name]",
	},
	cmd{
		name:  "about",
		desc:  "shows the requirement and description of a route",
		usage: "about [full_route]",
	},
	cmd{
		name:  "list",
		desc:  "lists all route that have been added to rubik server",
		usage: "list",
	},
	cmd{
		name:  "find",
		desc:  "finds the route that likely matches your argument",
		usage: "find [route]",
	},
	cmd{
		name:  "exit",
		desc:  "exit the okrubik repl",
		usage: "exit",
	},
}

// HelpCommand returns string to print for help command
func HelpCommand(args []string) string {
	skeleton := `  


	██████╗ ██╗   ██╗██████╗ ██╗██╗  ██╗
	██╔══██╗██║   ██║██╔══██╗██║██║ ██╔╝
	██████╔╝██║   ██║██████╔╝██║█████╔╝ 
	██╔══██╗██║   ██║██╔══██╗██║██╔═██╗ 
	██║  ██║╚██████╔╝██████╔╝██║██║  ██╗
	╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═╝	
										   	
   
Tooling Commands:

%s

REPL Commands:

%s
	`
	var cmdStr string
	var replStr string
	for _, c := range allCommands {
		cmdStr += fmt.Sprintf("%s 		 %s - %s\n", c.name, c.desc, c.usage)
	}
	for _, c := range replCommands {
		replStr += fmt.Sprintf("%s 		 %s - %s\n", c.name, c.desc, c.usage)
	}
	return fmt.Sprintf(skeleton, cmdStr, replStr)
}
