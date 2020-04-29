package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/endocrimes/keylight-go"
	"github.com/mitchellh/cli"
)

type DiscoverCommand struct {
	Meta
}

func (c *DiscoverCommand) Help() string {
	helpText := `
Usage: keylightctl discover [options]

 Discover all keylights that are currently available on the local network.

General Options:

  ` + generalOptionsUsage() + `

Discover Specific Options:

  -timeout <duration>
    Sets the maximum time to listen for accessories (default: 5s)

`
	return strings.TrimSpace(helpText)
}

func (f *DiscoverCommand) Synopsis() string {
	return "Discover keylights on the local network"
}

func (f *DiscoverCommand) Name() string { return "discover" }

func (c *DiscoverCommand) Run(args []string) int {
	c.UI = &cli.PrefixedUi{
		OutputPrefix: "  ",
		InfoPrefix:   "  ",
		ErrorPrefix:  "==> ",
		Ui:           c.UI,
	}

	var timeout string

	flags := c.Meta.FlagSet(c.Name(), FlagSetClient)
	flags.Usage = func() { c.UI.Output(c.Help()) }
	flags.StringVar(&timeout, "timeout", "5s", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if l := len(args); l != 0 {
		c.UI.Error("This command takes no arguments")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to parse timeout, err: %v", err))
		return 1
	}

	discovery, err := keylight.NewDiscovery()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to setup mDNS discovery, err: %v", err))
		return 1
	}

	discoveryCtx, cancelFn := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancelFn()

	go func(d keylight.Discovery) {
		d.Run(discoveryCtx)
	}(discovery)

	results := discovery.ResultsCh()

	c.UI.Info("Starting discovery")

	count := 0
	for a := range results {
		count++
		c.UI.Output(fmt.Sprintf("- %s", a.Name))
	}

	if count == 0 {
		c.UI.Error("Found no accessories during discovery")
		return 1
	}

	c.UI.Info(fmt.Sprintf("Found %d light(s) during discovery", count))

	return 0
}
