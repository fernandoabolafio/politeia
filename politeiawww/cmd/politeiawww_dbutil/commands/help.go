package commands

import "fmt"

type HelpCmd struct {
	Args struct {
		Topic string `positional-arg-name:"topic" description:"get information about available commands"`
	} `positional-args:"true"  required:"true"`
}

func (cmd *HelpCmd) Execute(args []string) error {

	switch cmd.Args.Topic {
	case "setadmin":
		fmt.Printf("%s\n", SetAdminCmdHelpMsg)
	default:
		fmt.Printf("invalid command\n")
	}

	return nil
}
