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
	{
		name:  "create",
		desc:  "creates a new rubik project",
		usage: "okrubik create",
	},
	{
		name:  "run",
		desc:  "runs project from your rubik workspace",
		usage: "okrubik run",
	},
	{
		name:  "x",
		desc:  "executes a defined command from your rubik workspace",
		usage: "okrubik x:test",
	},
	{
		name:  "gen",
		desc:  "generates boilerplate code for your rubik project",
		usage: "okrubik gen router",
	},
	{
		name:  "upgrade",
		desc:  "upgrades project modules || `self` option upgrade itself",
		usage: "okrubik upgrade [self]",
	},
}

var replCommands = []cmd{
	{
		name:  "select",
		desc:  "selects the project that you have initialized inside rubik.toml",
		usage: "select [project_name]",
	},
	{
		name:  "about",
		desc:  "shows the requirement and description of a route",
		usage: "about [full_route]",
	},
	{
		name:  "list",
		desc:  "lists all route that have been added to rubik server",
		usage: "list",
	},
	{
		name:  "find",
		desc:  "finds the route that likely matches your argument",
		usage: "find [route]",
	},
	{
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
`
	var cmdStr string
	// var replStr string
	for _, c := range allCommands {
		cmdStr += fmt.Sprintf("%s 		 %s - %s\n", c.name, c.desc, c.usage)
	}
	// for _, c := range replCommands {
	// 	replStr += fmt.Sprintf("%s 		 %s - %s\n", c.name, c.desc, c.usage)
	// }
	return fmt.Sprintf(skeleton, cmdStr)
}
