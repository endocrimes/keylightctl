package command

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/endocrimes/keylight-go"
	"github.com/jedib0t/go-pretty/table"
	"github.com/mitchellh/cli"
)

type DescribeCommand struct {
	Meta
}

func (c *DescribeCommand) Help() string {
	helpText := `
Usage: keylightctl describe [options]

 Describe the current state of the detected keylights

General Options:

  ` + generalOptionsUsage() + `

Describe Specific Options:

  -timeout <duration>
    Sets the maximum time to listen for accessories (default: 5s)

  -all
    Describe all keylights that are discovered within the timeout window

  -light <light-id>
    Modify the provided light ID. Can either be a full key light name, e.g:
    Elgato\ Key\ Light\ 111A, or a short ID, e.g: 111A. -light can be provided
    multiple times and all provided lights will be described.
`
	return strings.TrimSpace(helpText)
}

func (f *DescribeCommand) Synopsis() string {
	return "Describe the current state of the detected keylights"
}

func (f *DescribeCommand) Name() string { return "describe" }

func (c *DescribeCommand) Run(args []string) int {
	c.UI = &cli.PrefixedUi{
		OutputPrefix: "  ",
		InfoPrefix:   "  ",
		ErrorPrefix:  "==> ",
		Ui:           c.UI,
	}

	var timeout time.Duration
	var lights lightListFlags
	var allLights bool

	flags := c.Meta.FlagSet(c.Name(), FlagSetClient)
	flags.Usage = func() { c.UI.Output(c.Help()) }
	flags.DurationVar(&timeout, "timeout", 5*time.Second, "")
	flags.Var(&lights, "light", "")
	flags.BoolVar(&allLights, "all", false, "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if l := len(args); l != 0 {
		c.UI.Error("This command takes no arguments")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	if allLights && len(lights) != 0 {
		c.UI.Error("Cannot specify --all and --light")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	if !allLights && len(lights) == 0 {
		c.UI.Error("Must specify one of --all and --light")
		c.UI.Error(commandErrorText(c))
		return 1
	}

	discovery, err := keylight.NewDiscovery()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to setup mDNS discovery, err: %v", err))
		return 1
	}

	discoverer := lightDiscoverer{
		Discovery:      discovery,
		AllLights:      allLights,
		RequiredLights: lights,
	}

	discoveryCtx, cancelFn := context.WithTimeout(context.Background(), timeout)
	defer cancelFn()

	found, err := discoverer.Run(discoveryCtx)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to discover lights, err: %v", err))
		return 1
	}

	if len(found) == 0 {
		c.UI.Error("Found no matching lights during discovery")
		return 1
	}

	updateCtx, updateCancelFn := context.WithTimeout(context.Background(), 15*time.Second)
	defer updateCancelFn()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name", "Power State", "Brightness", "Temperature"})

	for idx, light := range found {
		opts, err := light.FetchLightGroup(updateCtx)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to fetch light options (%s), err: %v", light.Name, err))
			return 1
		}
		powerState := "off"
		if opts.Lights[0].On == 1 {
			powerState = "on"
		}

		t.AppendRows([]table.Row{
			{idx, light.Name, powerState, opts.Lights[0].Brightness, opts.Lights[0].Temperature},
		})
	}
	t.Render()

	return 0
}
