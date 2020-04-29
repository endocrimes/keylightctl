package main

import (
	"fmt"
	"os"

	"github.com/endocrimes/keylightctl/command"
	"github.com/endocrimes/keylightctl/version"
	"github.com/mattn/go-colorable"
	"github.com/mitchellh/cli"
	"golang.org/x/crypto/ssh/terminal"
)

func isColorEnabled(args []string) bool {
	noColor := false
	for _, arg := range args {
		// Check if color is set
		if arg == "-no-color" || arg == "--no-color" {
			noColor = true
		}
	}

	return !noColor
}

func main() {
	metaPtr := new(command.Meta)

	metaPtr.UI = &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      colorable.NewColorableStdout(),
		ErrorWriter: colorable.NewColorableStderr(),
	}

	isTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	args := os.Args
	color := isColorEnabled(args)

	// Only use colored UI if stdout is a tty, and not disabled
	if isTerminal && color {
		metaPtr.UI = &cli.ColoredUi{
			ErrorColor: cli.UiColorRed,
			WarnColor:  cli.UiColorYellow,
			InfoColor:  cli.UiColorGreen,
			Ui:         metaPtr.UI,
		}
	}

	commands := command.Commands(metaPtr)
	cli := &cli.CLI{
		Name:                       "keylightctl",
		Version:                    version.GetVersion().FullVersionNumber(true),
		Args:                       args[1:],
		Commands:                   commands,
		Autocomplete:               true,
		AutocompleteNoDefaultFlags: true,
		HelpFunc:                   cli.BasicHelpFunc("keylightctl"),
		HelpWriter:                 os.Stdout,
	}

	exitCode, err := cli.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(exitCode)
}
